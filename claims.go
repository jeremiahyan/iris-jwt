package jwt

import (
	"errors"
	"time"
)

var (
	// ErrExpired indicates that token is used after expiry time indicated in "exp" claim.
	ErrExpired = errors.New("token expired")
	// ErrNotValidYet indicates that token is used before time indicated in "nbf" claim.
	ErrNotValidYet = errors.New("token not valid yet")
	// ErrIssuedInTheFuture indicates that the "iat" claim is in the future.
	ErrIssuedInTheFuture = errors.New("token issued in the future")
)

// Claims holds the standard JWT claims (payload fields).
type Claims struct {
	// The opposite of the exp claim. A number representing a specific
	// date and time in the format “seconds since epoch” as defined by POSIX.
	// This claim sets the exact moment from which this JWT is considered valid.
	// The current time (see "t" argument of the `Verify` function)
	// must be equal to or later than this date and time.
	NotBefore int64 `json:"nbf,omitempty"`
	// A number representing a specific date and time (in the same
	// format as exp and nbf) at which this JWT was issued.
	IssuedAt int64 `json:"iat,omitempty"`
	// A number representing a specific date and time in the
	// format “seconds since epoch” as defined by POSIX6.
	// This claims sets the exact moment from which
	// this JWT is considered invalid. This implementation allow for a certain skew
	// between clocks (by considering this JWT to be valid for a few minutes after the expiration
	// date, modify the `Clock` variable).
	Expiry int64 `json:"exp,omitempty"`
	// A string representing a unique identifier for this JWT. This claim may be
	// used to differentiate JWTs with other similar content (preventing replays, for instance). It is
	// up to the implementation to guarantee uniqueness
	ID string `json:"jti,omitempty"`
	// A string or URI that uniquely identifies the party
	// that issued the JWT. Its interpretation is application specific (there is no central authority
	// managing issuers).
	Issuer string `json:"iss,omitempty"`
	// A string or URI that uniquely identifies the party
	// that this JWT carries information about. In other words, the claims contained in this JWT
	// are statements about this party. The JWT spec specifies that this claim must be unique in
	// the context of the issuer or, in cases where that is not possible, globally unique. Handling of
	// this claim is application specific.
	Subject string `json:"sub,omitempty"`
	// Either a single string or URI or an array of such
	// values that uniquely identify the intended recipients of this JWT. In other words, when this
	// claim is present, the party reading the data in this JWT must find itself in the aud claim or
	// disregard the data contained in the JWT. As in the case of the iss and sub claims, this claim
	// is application specific.
	Audience []string `json:"aud,omitempty"`
}

func validateClaims(t time.Time, claims Claims) error {
	now := t.Round(time.Second).Unix()

	if claims.NotBefore > 0 {
		if now < claims.NotBefore {
			return ErrNotValidYet
		}
	}

	if claims.IssuedAt > 0 {
		if now < claims.IssuedAt {
			return ErrIssuedInTheFuture
		}
	}

	if claims.Expiry > 0 {
		if now > claims.Expiry {
			return ErrExpired
		}
	}

	return nil
}

// Merge accepts two claim structs or maps
// and returns a flattened JSON result of both (no checks for duplicatations are maden).
//
// Usage:
//
//  claims := Merge(map[string]interface{}{"foo":"bar"}, Claims{
//    MaxAge: 15 * time.Minute,
//    Issuer: "an-issuer",
//  })
//  Sign(alg, key, claims)
//
// Merge is automatically called when:
//
//  Sign(alg, key, claims, MaxAge(time.Duration))
//  Sign(alg, key, claims, WithClaims(Claims{...}))
func Merge(claims interface{}, other interface{}) []byte {
	claimsB, err := Marshal(claims)
	if err != nil {
		return nil
	}

	otherB, err := Marshal(other)
	if err != nil {
		return nil
	}

	if len(otherB) == 0 {
		return claimsB
	}

	claimsB = claimsB[0 : len(claimsB)-1] // remove last '}'
	otherB = otherB[1:]                   // remove first '{'

	raw := append(claimsB, ',')
	raw = append(raw, otherB...)
	return raw
}
