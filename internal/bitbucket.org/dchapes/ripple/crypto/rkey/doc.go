// Package rkey, short for Ripple key, implements Ripple Account
// Families as documented at https://ripple.com/wiki/Account_Family.
// This includes the ability to decode a Ripple secret key (sXXX) into
// the family of private and public ECDSA keys and the associated Ripple
// addresses.
//
// Most types implement text marshaling (via encoding.TextMarshaler) to go to
// and from the Ripple base58 encoding.
// For example
//     s := new(FamilySeed);      s.UnmarshalText('sXXXX')
//     f := new(PublicGenerator); f.UnmarshalText('fXXXX')
//     p := new(AcctPrivateKey);  p.UnmarshalText('pXXXX')
//     a := new(AcctPublicKey);   a.UnmarshalText('aXXXX')
//
// GenerateSeed can be used to generate a new random FamilySeed and get the
// matching keys and Ripple address.
//
// The types are structured like the following pseudo-code:
//     FamilySeed struct {
//         Seed: 128 bit random number, encoded as sXXX
//         PrivateGenerator struct {
//             D
//             PublicGenerator struct {
//                  X, Y: encoded as fXXX (from the compressed point)
//             }
//         }
//     }
//     PrivateGenerator.Generate(int) -> AcctPrivateKey, encoded as pXXX
//     PublicGenerator.Generate(int)  -> AcctPublicKey,  encoded as aXXX
// Thus creating a FamilySeed also creates the generators.
// It's also possible to create and use just a PublicGenerator or just
// an AcctPublicKey without knowing the FamilySeed, PrivateGenerator, etc.
//
// TODO(dchapes): Currently all types implement Key which is an exported
// interface. This probably shouldn't be exported and ideally a
// better/cleaner implementation would be used.
package rkey
