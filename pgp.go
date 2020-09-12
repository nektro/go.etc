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

func PgpEncrypt(c *htp.Controller, message string, keys ...string) string {
	recs := []*openpgp.Entity{}
	for _, item := range keys {
		block, err := armor.Decode(strings.NewReader(item))
		c.AssertNilErr(err)
		c.Assert(block.Type == openpgp.PublicKeyType, "403: pgp block must be a public key")
		keyEnt, err := openpgp.ReadEntity(packet.NewReader(block.Body))
		c.AssertNilErr(err)
		recs = append(recs, keyEnt)
	}

	buf := new(bytes.Buffer)
	armWriter, err := armor.Encode(buf, "PGP MESSAGE", map[string]string{})
	c.AssertNilErr(err)
	pgpWriter, err := openpgp.Encrypt(armWriter, recs, nil, nil, nil)
	c.AssertNilErr(err)

	io.Copy(pgpWriter, strings.NewReader(message))
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
