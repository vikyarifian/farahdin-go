// Code generated from farahdin-react-native/app/pages/matrix-destiny.tsx by a one-off
// translation script; the expressions are copied verbatim (reduceNumber -> r,
// aPoint/bPoint/cPoint -> a/b/c) so the Go output matches the source app exactly.
// Edit by hand only together with matrix_test.go.

package domain

// matrixPoints returns the 32 points values of the destiny matrix.
func matrixPoints(a, b, c int) map[string]int {
	r := ReduceNumber
	return map[string]int{
		"apoint":  a,
		"bpoint":  b,
		"cpoint":  c,
		"dpoint":  r(a + b + c),
		"epoint":  r(a + b + c + r(a+b+c)),
		"fpoint":  r(a + b),
		"gpoint":  r(b + c),
		"hpoint":  r(r(a+b+c) + a),
		"ipoint":  r(c + r(a+b+c)),
		"jpoint":  r(r(a+b+c) + r(a+b+c+r(a+b+c))),
		"npoint":  r(c + r(a+b+c+r(a+b+c))),
		"lpoint":  r(r(r(a+b+c)+r(a+b+c+r(a+b+c))) + r(c+r(a+b+c+r(a+b+c)))),
		"mpoint":  r(r(r(r(a+b+c)+r(a+b+c+r(a+b+c)))+r(c+r(a+b+c+r(a+b+c)))) + r(c+r(a+b+c+r(a+b+c)))),
		"kpoint":  r(r(r(a+b+c)+r(a+b+c+r(a+b+c))) + r(r(r(a+b+c)+r(a+b+c+r(a+b+c)))+r(c+r(a+b+c+r(a+b+c))))),
		"qpoint":  r(r(c+r(a+b+c+r(a+b+c))) + c),
		"rpoint":  r(r(r(a+b+c)+r(a+b+c+r(a+b+c))) + r(a+b+c)),
		"spoint":  r(a + r(a+b+c+r(a+b+c))),
		"tpoint":  r(b + r(a+b+c+r(a+b+c))),
		"opoint":  r(a + r(a+r(a+b+c+r(a+b+c)))),
		"ppoint":  r(b + r(b+r(a+b+c+r(a+b+c)))),
		"upoint":  r(r(a+b) + r(b+c) + r(r(a+b+c)+a) + r(c+r(a+b+c))),
		"vpoint":  r(r(a+b+c+r(a+b+c)) + r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c)))),
		"wpoint":  r(r(a+r(a+b+c+r(a+b+c))) + r(a+b+c+r(a+b+c))),
		"xpoint":  r(r(b+r(a+b+c+r(a+b+c))) + r(a+b+c+r(a+b+c))),
		"f2point": r(r(a+b) + r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c)))),
		"f1point": r(r(a+b) + r(r(a+b)+r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c))))),
		"g2point": r(r(b+c) + r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c)))),
		"g1point": r(r(b+c) + r(r(b+c)+r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c))))),
		"i2point": r(r(c+r(a+b+c)) + r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c)))),
		"i1point": r(r(c+r(a+b+c)) + r(r(c+r(a+b+c))+r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c))))),
		"h2point": r(r(r(a+b+c)+a) + r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c)))),
		"h1point": r(r(r(a+b+c)+a) + r(r(r(a+b+c)+a)+r(r(a+b)+r(b+c)+r(r(a+b+c)+a)+r(c+r(a+b+c))))),
	}
}

// matrixYears returns the 56 years values of the destiny matrix.
func matrixYears(a, b, c int) map[string]int {
	r := ReduceNumber
	return map[string]int{
		"afpoint":  r(a + r(a+b)),
		"af1point": r(a + r(a+r(a+b))),
		"af2point": r(a + r(a+r(a+r(a+b)))),
		"af3point": r(r(a+r(a+b)) + r(a+r(a+r(a+b)))),
		"af4point": r(r(a+r(a+b)) + r(a+b)),
		"af5point": r(r(a+r(a+b)) + r(r(a+r(a+b))+r(a+b))),
		"af6point": r(r(r(a+r(a+b))+r(a+b)) + r(a+b)),
		"fbpoint":  r(r(a+b) + b),
		"fb1point": r(r(a+b) + r(r(a+b)+b)),
		"fb2point": r(r(a+b) + r(r(a+b)+r(r(a+b)+b))),
		"fb3point": r(r(r(a+b)+b) + r(r(a+b)+r(r(a+b)+b))),
		"fb4point": r(r(r(a+b)+b) + b),
		"fb5point": r(r(r(a+b)+b) + r(r(r(a+b)+b)+b)),
		"fb6point": r(r(r(r(a+b)+b)+b) + b),
		"bgpoint":  r(b + r(b+c)),
		"bg1point": r(b + r(b+r(b+c))),
		"bg2point": r(b + r(b+r(b+r(b+c)))),
		"bg3point": r(r(b+r(b+c)) + r(b+r(b+r(b+c)))),
		"bg4point": r(r(b+r(b+c)) + r(b+c)),
		"bg5point": r(r(b+r(b+c)) + r(r(b+r(b+c))+r(b+c))),
		"bg6point": r(r(r(b+r(b+c))+r(b+c)) + r(b+c)),
		"gcpoint":  r(r(b+c) + c),
		"gc1point": r(r(b+c) + r(r(b+c)+c)),
		"gc2point": r(r(b+c) + r(r(b+c)+r(r(b+c)+c))),
		"gc3point": r(r(r(b+c)+c) + r(r(b+c)+r(r(b+c)+c))),
		"gc4point": r(r(r(b+c)+c) + c),
		"gc5point": r(r(r(b+c)+c) + r(r(r(b+c)+c)+c)),
		"gc6point": r(r(r(r(b+c)+c)+c) + c),
		"cipoint":  r(c + r(c+r(a+b+c))),
		"ci1point": r(c + r(c+r(c+r(a+b+c)))),
		"ci2point": r(c + r(c+r(c+r(c+r(a+b+c))))),
		"ci3point": r(r(c+r(c+r(a+b+c))) + r(c+r(c+r(c+r(a+b+c))))),
		"ci4point": r(r(c+r(c+r(a+b+c))) + r(c+r(a+b+c))),
		"ci5point": r(r(c+r(c+r(a+b+c))) + r(r(c+r(c+r(a+b+c)))+r(c+r(a+b+c)))),
		"ci6point": r(r(r(c+r(c+r(a+b+c)))+r(c+r(a+b+c))) + r(c+r(a+b+c))),
		"idpoint":  r(r(c+r(a+b+c)) + r(a+b+c)),
		"id1point": r(r(c+r(a+b+c)) + r(r(c+r(a+b+c))+r(a+b+c))),
		"id2point": r(r(c+r(a+b+c)) + r(r(c+r(a+b+c))+r(r(c+r(a+b+c))+r(a+b+c)))),
		"id3point": r(r(r(c+r(a+b+c))+r(a+b+c)) + r(r(c+r(a+b+c))+r(r(c+r(a+b+c))+r(a+b+c)))),
		"id4point": r(r(r(c+r(a+b+c))+r(a+b+c)) + r(a+b+c)),
		"id5point": r(r(r(c+r(a+b+c))+r(a+b+c)) + r(r(r(c+r(a+b+c))+r(a+b+c))+r(a+b+c))),
		"id6point": r(r(r(r(c+r(a+b+c))+r(a+b+c))+r(a+b+c)) + r(a+b+c)),
		"dhpoint":  r(r(a+b+c) + r(r(a+b+c)+a)),
		"dh1point": r(r(a+b+c) + r(r(a+b+c)+r(r(a+b+c)+a))),
		"dh2point": r(r(a+b+c) + r(r(a+b+c)+r(r(a+b+c)+r(r(a+b+c)+a)))),
		"dh3point": r(r(r(a+b+c)+r(r(a+b+c)+a)) + r(r(a+b+c)+r(r(a+b+c)+r(r(a+b+c)+a)))),
		"dh4point": r(r(r(a+b+c)+r(r(a+b+c)+a)) + r(r(a+b+c)+a)),
		"dh5point": r(r(r(a+b+c)+r(r(a+b+c)+a)) + r(r(r(a+b+c)+r(r(a+b+c)+a))+r(r(a+b+c)+a))),
		"dh6point": r(r(r(r(a+b+c)+r(r(a+b+c)+a))+r(r(a+b+c)+a)) + r(r(a+b+c)+a)),
		"hapoint":  r(r(r(a+b+c)+a) + a),
		"ha1point": r(r(r(a+b+c)+a) + r(r(r(a+b+c)+a)+a)),
		"ha2point": r(r(r(a+b+c)+a) + r(r(r(a+b+c)+a)+r(r(r(a+b+c)+a)+a))),
		"ha3point": r(r(r(r(a+b+c)+a)+a) + r(r(r(a+b+c)+a)+r(r(r(a+b+c)+a)+a))),
		"ha4point": r(r(r(r(a+b+c)+a)+a) + a),
		"ha5point": r(r(r(r(a+b+c)+a)+a) + r(r(r(r(a+b+c)+a)+a)+a)),
		"ha6point": r(r(r(r(r(a+b+c)+a)+a)+a) + a),
	}
}

// matrixPurposes returns the 8 purposes values of the destiny matrix.
func matrixPurposes(a, b, c int) map[string]int {
	r := ReduceNumber
	return map[string]int{
		"skypoint":         r(b + r(a+b+c)),
		"earthpoint":       r(a + c),
		"perspurpose":      r(r(b+r(a+b+c)) + r(a+c)),
		"femalepoint":      r(r(b+c) + r(r(a+b+c)+a)),
		"malepoint":        r(r(a+b) + r(c+r(a+b+c))),
		"socialpurpose":    r(r(r(b+c)+r(r(a+b+c)+a)) + r(r(a+b)+r(c+r(a+b+c)))),
		"generalpurpose":   r(r(r(b+r(a+b+c))+r(a+c)) + r(r(r(b+c)+r(r(a+b+c)+a))+r(r(a+b)+r(c+r(a+b+c))))),
		"planetarypurpose": r(r(r(r(b+c)+r(r(a+b+c)+a))+r(r(a+b)+r(c+r(a+b+c)))) + r(r(r(b+r(a+b+c))+r(a+c))+r(r(r(b+c)+r(r(a+b+c)+a))+r(r(a+b)+r(c+r(a+b+c)))))),
	}
}

// matrixChakras returns the 21 chakras values of the destiny matrix.
func matrixChakras(a, b, c int) map[string]int {
	r := ReduceNumber
	return map[string]int{
		"sahphysics":   a,
		"ajphysics":    r(a + r(a+r(a+b+c+r(a+b+c)))),
		"vishphysics":  r(a + r(a+b+c+r(a+b+c))),
		"anahphysics":  r(r(a+r(a+b+c+r(a+b+c))) + r(a+b+c+r(a+b+c))),
		"manphysics":   r(a + b + c + r(a+b+c)),
		"svadphysics":  r(c + r(a+b+c+r(a+b+c))),
		"mulphysics":   c,
		"sahenergy":    b,
		"ajenergy":     r(b + r(b+r(a+b+c+r(a+b+c)))),
		"vishenergy":   r(b + r(a+b+c+r(a+b+c))),
		"anahenergy":   r(r(b+r(a+b+c+r(a+b+c))) + r(a+b+c+r(a+b+c))),
		"manenergy":    r(a + b + c + r(a+b+c)),
		"svadenergy":   r(r(a+b+c) + r(a+b+c+r(a+b+c))),
		"mulenergy":    r(a + b + c),
		"sahemotions":  r(a + b),
		"ajemotions":   r(r(a+r(a+r(a+b+c+r(a+b+c)))) + r(b+r(b+r(a+b+c+r(a+b+c))))),
		"vishemotions": r(r(a+r(a+b+c+r(a+b+c))) + r(b+r(a+b+c+r(a+b+c)))),
		"anahemotions": r(r(r(a+r(a+b+c+r(a+b+c)))+r(a+b+c+r(a+b+c))) + r(r(b+r(a+b+c+r(a+b+c)))+r(a+b+c+r(a+b+c)))),
		"manemotions":  r(r(a+b+c+r(a+b+c)) + r(a+b+c+r(a+b+c))),
		"svademotions": r(r(r(a+b+c)+r(a+b+c+r(a+b+c))) + r(c+r(a+b+c+r(a+b+c)))),
		"mulemotions":  r(c + r(a+b+c)),
	}
}
