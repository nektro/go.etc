package etc

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/nektro/go.etc/htp"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func PgpEncrypt(c *htp.Controller, keyText, message string) string {
	block, err := armor.Decode(strings.NewReader(keyText))
	c.AssertNilErr(err)
	c.Assert(block.Type == openpgp.PublicKeyType, "403: pgp block must be a public key")

	keyEnt, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	c.AssertNilErr(err)

	buf := new(bytes.Buffer)

	armWriter, err := armor.Encode(buf, "PGP MESSAGE", map[string]string{})
	c.AssertNilErr(err)

	pgpWriter, err := openpgp.Encrypt(armWriter, []*openpgp.Entity{keyEnt}, nil, nil, nil)
	c.AssertNilErr(err)

	io.Copy(pgpWriter, bytes.NewReader([]byte(message)))
	pgpWriter.Close()
	armWriter.Close()

	return string(buf.Bytes())
}

func PgpPubKeyFingerprint(c *htp.Controller, keyText string) string {
	block, err := armor.Decode(strings.NewReader(keyText))
	c.AssertNilErr(err)
	c.Assert(block.Type == openpgp.PublicKeyType, "403: pgp block must be a public key")
	keyEnt, err := openpgp.ReadEntity(packet.NewReader(block.Body))
	c.AssertNilErr(err)
	finger := fmt.Sprintf("%X", keyEnt.PrimaryKey.Fingerprint[:])
	return finger
}
