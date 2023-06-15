// Desc: Configuration file for DAS
const dasConfig = {
  assumedHonest: process.env.HONEST,
  members: (process.env.URLS || "").split(',').map(url => ({ url }))
};

module.exports = { dasConfig };
