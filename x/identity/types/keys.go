package types

const (
	ModuleName = "identity"
	StoreKey   = ModuleName
)

var KeyPrefixAlias = []byte("alias/")

func AliasKey(passkeyPubkey string) []byte {
	return append(KeyPrefixAlias, []byte(passkeyPubkey)...)
}
