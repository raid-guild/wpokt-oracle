import type { Abi } from "viem";

import OmniToken from "./OmniToken.json";
import MintController from "./wPOKTMintController.json";
import Mailbox from "./Mailbox.json";
import WarpISM from "./WarpISM.json";
import AccountFactory from "./AccountFactory.json";
import Account from "./Account.json";
import Multicall3 from "./Multicall3.json";


export const OmniTokenAbi: Abi = OmniToken as Abi;
export const MintControllerAbi: Abi = MintController as Abi;
export const MailboxAbi: Abi = Mailbox as Abi;
export const WarpISMAbi: Abi = WarpISM as Abi;
export const AccountFactoryAbi: Abi = AccountFactory as Abi;
export const AccountAbi: Abi = Account as Abi;
export const Multicall3Abi: Abi = Multicall3 as Abi;
