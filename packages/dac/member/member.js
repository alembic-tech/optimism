require("dotenv").config();
const express = require("express");
const formidableMiddleware = require("express-formidable");
const member = express();
member.use(formidableMiddleware());
const bls = require("@noble/curves/bls12-381").bls12_381;
const utils = require("@noble/curves/abstract/utils");
const crypto = require("./crypto");
const dataSchema = require("./models/Data");
const { dbInstance, connectDB } = require("./db");

const { memberConfig } = require("./config.js");

connectDB();
member.post("/signCert", async (req, res) => {
  let { data /*timeout, sig*/ } = req.fields;
  // timeout = new Uint8Array(timeout);

  if (data) {
    const msg = data; /*+ timeout*/

    // Verify if the batcher signature is valid
    // let isValidSequencerSig = crypto.verifySignature(
    //   msg,
    //   sig,
    //   memberConfig.publicKey
    // );

    let isValidSequencerSig = true;
    console.log("-------isValidSequencerSig", isValidSequencerSig);
    if (isValidSequencerSig) {
      // Sign the data
      const dataHash = crypto.hashMsg(msg);
      const sig = utils.bytesToHex(bls.sign(dataHash, memberConfig.privateKey));

      // Save the data to the database
      if (!dbInstance.models["Data"]) dbInstance.model("Data", dataSchema);
      const newData = dbInstance.model("Data")({
        data,
        dataHash,
        // timeout,
      });
      await newData.save();

      const certDetails = {
        sig,
        dataHash,
      };

      console.log("-------certDetails", certDetails);
      res.status(200).json(certDetails);
    } else {
      res.status(400).json({ message: "Invalid sequencer/batcher signature" });
    }
  } else {
    res.status(400).json({ message: "Missing data or timeout" });
  }
});

member.get("/batch/:dataHash", async (req, res) => {
  const { dataHash } = req.params;
  if (!dbInstance.models["Data"]) dbInstance.model("Data", dataSchema);
  const data = await dbInstance.model("Data").findOne({ dataHash: dataHash });

  if (data) {
    res.status(200).json(data);
  } else {
    res.status(400).json({ message: "Data not found" });
  }
});

member.all("*", function (req, res) {
  res.json({ message: "Not found" });
});

member.listen(process.env.PORT, () => {
  console.log(`Committee member on port ${process.env.PORT}`);
});

module.exports = member;
