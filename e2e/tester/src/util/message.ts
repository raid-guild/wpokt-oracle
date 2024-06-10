import {
  Hex, encodePacked, encodeAbiParameters
} from "viem";
import { MessageBody, MessageContent } from "../types";


function addressHexToBytes32(address: Hex): Hex {
  return `0x${address.slice(2).padStart(64, '0')}`;
}

function formatMessage(
  version: number,
  nonce: number,
  originDomain: number,
  sender: Hex,
  destinationDomain: number,
  recipient: Hex,
  messageBody: Hex
): Hex {
  return encodePacked(
    ['uint8', 'uint32', 'uint32', 'bytes32', 'uint32', 'bytes32', 'bytes'],
    [version, nonce, originDomain, addressHexToBytes32(sender), destinationDomain, addressHexToBytes32(recipient), messageBody]
  );
}

function formatMessageBody(
  recipient: Hex,
  amount: bigint,
  sender: Hex
): Hex {
  return encodeAbiParameters(
    [{ type: 'address', name: 'recipient' }, { type: 'uint256', name: 'amount' }, { type: 'address', name: 'sender' }],
    [recipient, amount, sender]
  );
}

const encodeMessageBody = (messageBody: MessageBody): Hex => {
  return formatMessageBody(
    messageBody.recipient_address,
    BigInt(messageBody.amount),
    messageBody.sender_address
  );
}

export const encodeMessage = (message: MessageContent): Hex => {

  const messageBodyHex = encodeMessageBody(
    message.message_body
  );

  return formatMessage(
    message.version,
    message.nonce,
    message.origin_domain,
    message.sender,
    message.destination_domain,
    message.recipient,
    messageBodyHex
  );

}

export const concatHex = (values: readonly Hex[]): Hex => {
  return `0x${(values as Hex[]).reduce(
    (acc, x) => acc + x.replace('0x', ''),
    '',
  )}`
}
