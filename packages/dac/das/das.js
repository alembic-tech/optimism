require("dotenv").config();
const express = require("express");
const formidableMiddleware = require("express-formidable");
const das = express();
das.use(formidableMiddleware());
const axios = require("axios");
const bls = require("@noble/curves/bls12-381").bls12_381;
const { dasConfig } = require("./config.js");

const encodeHex = (arr) => Buffer.from(arr).toString('hex');

// FIXME: WRAP PROMISES !!!!
// FIXME: WRAP PROMISES !!!!
// FIXME: WRAP PROMISES !!!!
// FIXME: WRAP PROMISES !!!!
// FIXME: WRAP PROMISES !!!!
// FIXME: WRAP PROMISES !!!!
// FIXME: WRAP PROMISES !!!!

das.post("/batch", async (req, res) => {
  const { data } = req.fields;
  const body = { data };

  const results = await Promise.allSettled(
    dasConfig.members.map(
      (member) =>
        axios.post(
          `${member.url}/batch`,
          body
        )
    )
  );

  results.filter((result) => result.status === 'rejected').forEach((result) => {
    console.error(result.reason);
  });

  const fulfilled = results.filter((result) => result.status === 'fulfilled');
  if (!fulfilled.length) {
    res.status(500).send();
    return;
  }
  // FIXME: we should check that dataHash is valid (or at least consistent) ?
  const dataHash = fulfilled[0].value.data.data_hash

  const [signatures, publicKeys] = fulfilled.reduce(
    ([signatures, publicKeys], { value }) => [[...signatures, value.data.signature], [...publicKeys, value.data.public_key]],
    [[], []],
  );

  const aggregatedSignature = bls.aggregateSignatures(
    signatures.map((encoded) => bls.G2.ProjectivePoint.fromHex(encoded))
  );

  const response = {
    data_hash: dataHash,
    signature: encodeHex(aggregatedSignature.toRawBytes(false)),
    public_keys: publicKeys,
    signatures,
  }

  console.log(response)
  res.status(200).json(response);
});

das.get("/batch/:dataHash", async (req, res) => {
  const { dataHash } = req.params;
  let promises = [];
  const daMembers = dasConfig.members.length;
  for (let i = 0; i < daMembers; i++) {
    let promise = axios.get(
      `${dasConfig.members[i].url}/batch/${dataHash}`,
    );
    promises.push(promise);
  }
  const result = await Promise.any(promises);
  const data = result.data;
  res.status(200).json({ data });
});

const port = process.env.PORT || '3000'
das.listen(port, () => {
  console.log(`DAS on port ${port}`);
});

module.exports = das;
