package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/repository"
)

// UserStore is the persistence the profile use cases need.
type UserStore interface {
	ByID(ctx context.Context, id int64) (*domain.User, error)
	ByGoogleSub(ctx context.Context, sub string) (*domain.User, error)
	FirstByEmail(ctx context.Context, email string) (*domain.User, error)
	Insert(ctx context.Context, u *domain.User) (int64, error)
	LinkIdentity(ctx context.Context, id int64, username, fullname, sub, imageURL string) error
	SetImage(ctx context.Context, id int64, imageURL string) error
	UpdateProfile(ctx context.Context, id int64, fullname, birthday, birthplace, gender, zodiac string) error
	UpdateBirthday(ctx context.Context, id int64, birthday, zodiac string) error
}

// Profiles implements sign-up/sign-in user resolution and profile editing.
type Profiles struct {
	Users UserStore
	Now   func() time.Time // in the app's time zone
}

// SignIn ports convex/users.ts createUser (formerly run by the Clerk
// "user.created" webhook): match by provider id, then by email, else insert.
func (p *Profiles) SignIn(ctx context.Context, id domain.Identity) (*domain.User, error) {
	if id.Subject == "" || id.Email == "" {
		return nil, errors.New("identity without subject or email")
	}
	u, err := p.Users.ByGoogleSub(ctx, id.Subject)
	if err == nil {
		// createUser returned early here; only the photo URL is refreshed.
		if id.ImageURL != "" && id.ImageURL != u.ImageURL {
			if err := p.Users.SetImage(ctx, u.ID, id.ImageURL); err != nil {
				return nil, err
			}
			u.ImageURL = id.ImageURL
		}
		return u, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	u, err = p.Users.FirstByEmail(ctx, id.Email)
	if err == nil {
		if err := p.Users.LinkIdentity(ctx, u.ID, id.Username(), id.Fullname(), id.Subject, id.ImageURL); err != nil {
			return nil, err
		}
		return p.Users.ByID(ctx, u.ID)
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	nu := &domain.User{
		Username:  id.Username(),
		Fullname:  id.Fullname(),
		Email:     id.Email,
		Birthday:  domain.FormatDate(p.Now()), // createUser defaulted to "today"
		GoogleSub: id.Subject,
		ImageURL:  id.ImageURL,
	}
	newID, err := p.Users.Insert(ctx, nu)
	if err != nil {
		return nil, err
	}
	return p.Users.ByID(ctx, newID)
}

// Genders are the options of the Edit Profile radio group.
var Genders = []string{"Male", "Female"}

// ValidateProfile checks an edit before it is saved. The source app did not
// validate; these bounds only reject input its pickers could not produce.
func ValidateProfile(in domain.ProfileUpdate, today time.Time, lang string) map[string]string {
	id := lang == "ID"
	errs := map[string]string{}
	if utf8.RuneCountInString(in.Fullname) > 100 {
		errs["name"] = pick(id, "Nama maksimal 100 karakter.", "Name must be at most 100 characters.")
	}
	if utf8.RuneCountInString(in.Birthplace) > 200 {
		errs["birthplace"] = pick(id, "Alamat maksimal 200 karakter.", "Location must be at most 200 characters.")
	}
	if in.Birthday.IsZero() || in.Birthday.Before(domain.MinBirthDate) || in.Birthday.After(today) {
		errs["birthday"] = pick(id, "Tanggal lahir tidak valid.", "Birthdate is invalid.")
	}
	valid := false
	for _, g := range Genders {
		if in.Gender == g {
			valid = true
		}
	}
	if !valid {
		errs["gender"] = pick(id, "Pilih jenis kelamin.", "Choose a gender.")
	}
	return errs
}

// UpdateProfile ports users.updateUser as called from Edit Profile. Username
// and email are not editable (the source sent them back unchanged), and the
// zodiac is always derived from the birthday on the server.
func (p *Profiles) UpdateProfile(ctx context.Context, u *domain.User, in domain.ProfileUpdate, lang string) (map[string]string, error) {
	in.Fullname = strings.TrimSpace(in.Fullname)
	in.Birthplace = strings.TrimSpace(in.Birthplace)
	if errs := ValidateProfile(in, p.Now(), lang); len(errs) > 0 {
		return errs, nil
	}
	birthday := domain.FormatDate(in.Birthday)
	return nil, p.Users.UpdateProfile(ctx, u.ID, in.Fullname, birthday, in.Birthplace, in.Gender, domain.Zodiac(birthday))
}

// SyncBirthday ports the Horoscope/Primbon side effect that saved a changed
// birthday (and its zodiac) to the profile.
func (p *Profiles) SyncBirthday(ctx context.Context, u *domain.User, birthday time.Time) error {
	b := domain.FormatDate(birthday)
	if u.Birthday == b {
		return nil
	}
	if err := p.Users.UpdateBirthday(ctx, u.ID, b, domain.Zodiac(b)); err != nil {
		return err
	}
	u.Birthday, u.Zodiac = b, domain.Zodiac(b)
	return nil
}
