package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Translator ports utils/Translate.ts.
type Translator struct {
	Client *Client
	Base   string
}

// Translate returns one entry per translated sentence, as the source app did
// (data.sentences[].trans). Like the source, "&" is replaced with "dan"/"and"
// before sending; unlike the source, the text is URL-encoded properly.
func (t *Translator) Translate(ctx context.Context, text, from, to string) ([]string, error) {
	amp := "and"
	if from == "id" {
		amp = "dan"
	}
	q := url.Values{}
	q.Set("client", "gtx")
	q.Set("sl", from)
	q.Set("tl", to)
	q.Set("dj", "1")
	q.Set("dt", "t")
	q.Set("ie", "UTF-8")
	q.Set("q", strings.ReplaceAll(text, "&", amp))

	body, err := t.Client.Get(ctx, t.Base+"/translate_a/single?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("translate: %w", err)
	}
	var data struct {
		Sentences []struct {
			Trans *string `json:"trans"`
		} `json:"sentences"`
	}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, fmt.Errorf("translate: decode: %w", err)
	}
	out := make([]string, 0, len(data.Sentences))
	for _, s := range data.Sentences {
		if s.Trans != nil {
			out = append(out, *s.Trans)
		}
	}
	return out, nil
}
