import { formatUnits } from "viem";
import { findNodes } from "./util/mongodb";
import cosmos from "./util/cosmos";
import ethereum from "./util/ethereum";
import { config } from "./util/config";
// import { mintFlow } from "./flows/mint";
// import { burnFlow } from "./flows/burn";

const init = async () => {
  const nodes = await findNodes();

  console.log("Number of nodes:", nodes.length);

  const cosmosAddress = await cosmos.getAddress();

  console.log("Pocket address:", cosmosAddress);

  console.log(
    "Pocket network:",
    config.cosmos_network.chain_name,
    "with chain ID",
    config.cosmos_network.chain_id,
    "at",
    config.cosmos_network.rpc_url
  );

  const cosmosBalance = await cosmos.getBalance(cosmosAddress);

  console.log("Pocket balance:", formatUnits(cosmosBalance, 6), "POKT");


  console.log("Ethereum networks:");

  console.log("Number of networks:", config.ethereum_networks.length);

  for (const network of config.ethereum_networks) {

    const ethAddress = await ethereum.getAddress(network.chain_id);

    console.log("Ethereum address:", ethAddress);

    console.log(
      "Ethereum network:",
      network.chain_name,
      "with chain ID",
      network.chain_id,
      "at",
      network.rpc_url
    );

    const ethBalance = await ethereum.getBalance(network.chain_id, ethAddress);

    console.log("Ethereum balance:", formatUnits(ethBalance, 18), "ETH");
  }
};

before(async () => {
  await init();
  console.log("\n");
});

describe("E2E tests", async () => {
  // describe("Mint flow", mintFlow);
  // describe("Burn flow", burnFlow);

  describe("Basic", async () => {
    it("should pass", async () => {
      console.log("Basic test");
    });
  });

});
