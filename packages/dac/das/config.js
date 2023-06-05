// Desc: Configuration file for DAS
const dasConfig = {
  assumedHonest: process.env.HONEST,
  members: [
    {
      port: Number(process.env.PORT_MEMBER) + 1,
      pubKey: process.env.PUBKEY1,
    },
  ],
};

module.exports = { dasConfig };
