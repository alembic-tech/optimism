const mongoose = require("mongoose");
const { Schema } = mongoose;

const dataSchema = new Schema({
  dataHash: String,
  data: Array,
  // timeout: Number,
});

module.exports = dataSchema;
