require("dotenv").config();
const express = require("express");
const formidableMiddleware = require("express-formidable");
const das = express();
das.use(formidableMiddleware());
const axios = require("axios");
const bls = require("@noble/curves/bls12-381").bls12_381;
const utils = require("@noble/curves/abstract/utils");
const { dasConfig } = require("./config.js");
const allEqual = (arr) => arr.every((v) => v === arr[0]);

das.post("/batch", async (req, res) => {
  const { data /*timeout, sig*/ } = req.fields;

  // Distribute the data to the DA members
  let promises = [];
  const daMembers = dasConfig.members.length;
  const body = { data /* timeout, sig*/ };

  // distribute the data to the DA members
  for (let i = 0; i < daMembers; i++) {
    try {
      let promise = axios.post(
        `http://member:${dasConfig.members[i].port}/signCert`,
        body
      );
      promises.push(promise);
    } catch (error) {
      console.log(error);
    }
  }

  // Collect the signatures from the DA members
  let sigs = [],
    dataHashes = [],
    signersIndex = [];
  const results = await Promise.allSettled(promises);

  for (const result of results) {
    if (result.status === "fulfilled") {
      // Check individual committee signatures
      // const pubKeys = dasConfig.members.map((member) => member.publicKey);

      const { dataHash, sig } = result.value.data;
      sigs.push(utils.hexToBytes(sig));
      dataHashes.push(dataHash);
      // signersIndex.push("result.value.data.index");
    }
  }

  // Check Anytrust assumption
  const minRequiredSigs = daMembers - dasConfig.assumedHonest + 1;
  if (allEqual(dataHashes) && sigs.length >= minRequiredSigs) {
    const dataHash = dataHashes[0];

    // Aggregate the signatures and pubKeys
    const aggSignature = bls.aggregateSignatures(sigs);
    // const aggPubKey = bls.aggregatePublicKeys(pubKeys);

    // Pre check the aggregated signature
    // const isValidAggSignature = crypto.verifySignature(
    //   data + timeout,
    //   aggSignature,
    //   aggPubKey
    // );

    let isValidAggSignature = true;

    console.log("---------dataHash", dataHash);
    if (isValidAggSignature) {
      res.status(200).json({ dataHash /*signersIndex, aggSignature*/ });
    } else {
      res.status(400).json({ message: "Invalid agg signatures" });
    }
  } else {
    res.status(400).json({ message: "Error" });
  }
});

das.get("/batch/:dataHash", async (req, res) => {
  const { dataHash } = req.params;
  let promises = [];
  const daMembers = dasConfig.members.length;
  for (let i = 0; i < daMembers; i++) {
    let promise = axios.get(
      `http://member:${dasConfig.members[i].port}/batch/${dataHash}`
    );
    promises.push(promise);
  }
  const result = await Promise.any(promises);
  const data = result.data;
  res.status(200).json({ data });
});

das.all("*", function (req, res) {
  res.json({ message: "Not found" });
});

das.listen(process.env.PORT, () => {
  console.log(`DAS on port ${process.env.PORT}`);
});

module.exports = das;
