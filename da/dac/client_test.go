package dac

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestClient(t *testing.T) {
   dataHashStr := "340047ffe6c2c4af811ce86f2d0df0b73d7b1f30bcc285a6f06c2d3c6a8b2efa"
   signatureStr := "11e9bb0aea102d6db81ae001dc6915b2162f99baa7d6bc3309397410a7edea77b519898cf5475d044c5b1cc21fcf106310675afc0f64d28ab980c1a565567681347b6263b91c12ecb47545235b94fcf42732f52a56281d4586c6f830c595243f0da2a223b18b440ad70600aa2b5648ea5a86a813491dc29834f2aefd6e1b6ad78be85825249f13a7ac8b052c03268b8d05e494552b50169a368c753d8c9503be0faed1f65c82ca9b8e171845e83bb11186c5524d27fbb10f1fe318d36f54b988"

   publicKeys := []string{
     "10c36f69c5f73a0ae95fa1768e68a58973d0a3a61f1e9bf889050217388ebb24c57341fb5528b8f2b6138d5149d88c611003f241a22da86d76e15cdfdc06d6ea86845d5f662e3209044716add654d98aa6a9c99632b2d647ac280e36d9da5756",
     "1866562f34d7339fc7831b4b6a47defe714007f4720d03849f2e90b0ee8330d56d76e0957f500fe76f24e6a730016bcf091e3ba87744ceefd90503b120550eb9554f5683b739fb1cbdd6e56ac7b89288e9f35ce7e32d40008f347b21e036602c",
     "1830f58c42f446a5659c08a5c39a2734cdbca673007076612384387d3f79e6c44c2427012cac0b011c698b9c0d5834920a7ec51884e4f7b43d73a683d1038c1b940dd4c1aa69bd097419b8ed69ff9447e44c9da1e8ea38ac916655260a42b14e",
   }

   signaturesStr := []string{
     "02a55c9929cfb2e4087f591fc63fb0d09be473f1d595560a1a05f6a1320cd1bec8c38eb50e2b4b9f852108fa1adae483081e9632045f385213b95083654dc6c7a6dd647d795cd1d947827a0767300868cb9455e4bd288b9714da1cf575c1954900c24ef5ec3af346ef1a6872daaed1a5a2dbb77a2e2bc9eaf21280e938f04c90f98d06dd1b308e06775762f7c0910bab18fc6cb9f8c5f844c0c07daa9728563d7092abf9eef11db259fce88b688b6d9d27be5e930371cabac7799e37d774daa5",
     "19bf45e8daf66aa1bb2f41f70c37bab068eeb38b8b50548694549ceff56bf0c8d06fe5ee8e3c5c691ea8d757ae096b0c0e5291375f42fc2a46ad2c2423af60df7d49f20b8fec0f3cf36614b613f2c4b761edb683d6027923d6752e3c487d3f2d046037de1b4ca10903826a32650ecdfb68c6b342cf354c5727cd3de3e4e4f06c5f3e00e6ff340dbd7e46fc40c373c2ca163da73147c68bff541a2dfb0118f58a477166f1473c772a3eb32d99a0b80ea73ed014426bf6c9371d85b8eab37724a6",
     "00c33ec778c35b4577b4e5a4b40cd1d26bd8b6c9d88ccecc412d0c75cbd5bebfee08f95c4792d52d939ada4c0eb3451f19e760f4b474ef0bf5f75cefdb3a1694fb2d29a8e33c513bcca30bcb22bd6bf4c035b0193761f014993808fe988569a3032003fad1ac1d259ab7c29ae161d55897a24188f5dd0b3418ad2b87dc36d1c23ddf8b35c7a2934721e0e79c3fe21219017c64d5d0f3248eff87dd0290b5f61e6a0f52f82cbf73fb7d53b94a2b75640e4d4e2c384d674e0cfdddf309b5e757dc",
   }

   keyset, err := NewKeySetFromString(publicKeys)
   if err != nil {
     t.Fatal(err)
   }
   
   dataHash, _ := hex.DecodeString(dataHashStr)
   signature, _ := hex.DecodeString(signatureStr)

   c := NewClient("", (common.Address{}), keyset).(*client)
   fmt.Println(c.verifySignature(dataHash, keyset, signature))

   signatures := make([]Signature, len(signaturesStr))

   for i, str := range signaturesStr {
     b, _ := hex.DecodeString(str)
     signatures[i], err = NewSignature(b)
     if err != nil {
       t.Fatal(err)
     }
   }

   agg := AggregateSignatures(signatures)
   _ = agg

   signer, err := NewSigner("0x39bfcae8591588ef01774d3a5003d3a5b5c95a00b2142b20b217eedaeb124f63")
   if err != nil {
     t.Fatal(err)
   }
   sig, err := signer.Sign(dataHash)
   if err != nil {
     t.Fatal(err)
   }

   
   fmt.Println(signer.GetPublicKey().VerifyMessage(dataHash, sig.ToBytes()))
}
