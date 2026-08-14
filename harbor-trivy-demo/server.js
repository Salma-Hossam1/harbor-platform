const express = require("express");

const app = express();
const PORT = 7800;

app.get("/", (req, res) => {
  res.send("Hello Harbor");
});

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});