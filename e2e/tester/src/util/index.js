
function addressHexToBytes32(address) {
  return `0x${address.slice(2).padStart(64, '0')}`;
}

console.log(addressHexToBytes32('0xd8c0ba27c0b9f8682bd896cabf30b9f57a5631b3'));
console.log(addressHexToBytes32('0x5fc8d32690cc91d4c39d9d3abcbd16989f875707'));

