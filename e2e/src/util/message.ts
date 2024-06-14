import {
  Hex, encodePacked, encodeAbiParameters,
  decodeAbiParameters
} from "viem";
import { Long } from "mongodb";
import { MessageBody, MessageContent } from "../types";


export function addressHexToBytes32(address: Hex): Hex {
  return `0x${address.slice(2).padStart(64, '0')}`.toLowerCase() as Hex;
}

export function bytes32ToAddressHex(bytes32: Hex): Hex {
  if (bytes32.length !== 66) {
    throw new Error("Invalid bytes32 length");
  }
  return `0x${bytes32.slice(26)}`.toLowerCase() as Hex;
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

export function formatMessageBody(
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
    BigInt(messageBody.amount.toString()),
    messageBody.sender_address
  );
}

const decodeMessageBody = (messageBody: Hex): MessageBody => {
  const decoded = decodeAbiParameters(
    [{ type: 'address', name: 'recipient' }, { type: 'uint256', name: 'amount' }, { type: 'address', name: 'sender' }],
    messageBody
  );

  return {
    recipient_address: decoded[0].toLowerCase() as Hex,
    amount: new Long(decoded[1].toString()),
    sender_address: decoded[2].toLowerCase() as Hex,
  };
}

export const encodeMessage = (message: MessageContent): Hex => {

  const messageBodyHex = encodeMessageBody(
    message.message_body
  );

  return formatMessage(
    message.version,
    message.nonce.toNumber(),
    message.origin_domain.toNumber(),
    message.sender,
    message.destination_domain.toNumber(),
    message.recipient,
    messageBodyHex
  );

}

export function decodeMessage(encodedMessage: Hex): MessageContent {
  const data = Buffer.from(encodedMessage.slice(2), 'hex');

  if (data.length !== 173) {
    throw new Error("Invalid data length");
  }

  let offset = 0;
  const version = data.readUInt8(offset);
  offset += 1;

  const nonce = data.readUInt32BE(offset);
  offset += 4;

  const originDomain = data.readUInt32BE(offset);
  offset += 4;

  const senderBytes = data.subarray(offset, offset + 32);
  const sender = bytes32ToAddressHex(`0x${senderBytes.toString('hex')}`);
  offset += 32;

  const destinationDomain = data.readUInt32BE(offset);
  offset += 4;

  const recipientBytes = data.subarray(offset, offset + 32);
  const recipient = bytes32ToAddressHex(`0x${recipientBytes.toString('hex')}`);
  offset += 32;

  const messageBody = data.subarray(offset).toString('hex');

  const message_body = decodeMessageBody(`0x${messageBody}`);

  return {
    version,
    nonce: new Long(nonce),
    origin_domain: new Long(originDomain),
    sender,
    destination_domain: new Long(destinationDomain),
    recipient,
    message_body,
  };
}
