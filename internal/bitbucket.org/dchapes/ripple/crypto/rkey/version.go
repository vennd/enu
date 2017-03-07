package rkey

// version is the version or type of an ecoded Ripple base58 value/key.
type version byte

const (
	verNone            version = 1
	verFamilySeed              = 33 // sXXX
	verFamilyGenerator         = 41 // fXXX
	verAcctPrivate             = 34 // pXXX
	verAcctPublic              = 35 // aXXX
	verAcctId                  = 0  // rXXX
	verNodePublic              = 28 // nXXX, not implemented here
	verNodePrivate             = 32 // pXXX, not implemented here
)

var payloadSize = map[version]int8{
	verFamilySeed:      16,
	verFamilyGenerator: 33,
	verAcctPrivate:     32,
	verAcctPublic:      33,
	verAcctId:          20,
	verNodePublic:      33,
	verNodePrivate:     32,
}

func (v version) valid() bool {
	_, ok := payloadSize[v]
	return ok
}

func (v version) PayloadSize() int {
	if sz, ok := payloadSize[v]; ok {
		return int(sz)
	}
	panic("rkey: unknown version")
}
