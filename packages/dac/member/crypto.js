const bls = require("@noble/curves/bls12-381").bls12_381;
const keccak256 = require("keccak256");
const utils = require("@noble/curves/abstract/utils");

const verifySignature = (msg, sig, pubKey) => {
  const msgHash = keccak256(msg).toString("hex");
  pubKey = utils.hexToBytes(pubKey);
  console.log(sig);
  const isValid = bls.verify(sig, msgHash, pubKey);
  return isValid;
};

const hashMsg = (msg) => keccak256(msg).toString("hex");

module.exports = { verifySignature, hashMsg };
