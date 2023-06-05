const mongoose = require("mongoose");

const connectDB = async () =>
  await mongoose.connect(process.env.MONGO_DB, {
    useNewUrlParser: true,
    useUnifiedTopology: true,
  });

const dbInstance = mongoose.connection.useDb(process.env.MEMBER, {
  useCache: true,
});

module.exports = { connectDB, dbInstance };
