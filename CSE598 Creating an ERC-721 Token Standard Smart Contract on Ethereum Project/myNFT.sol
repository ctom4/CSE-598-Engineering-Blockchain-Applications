// SPDX-License-Identifier: MIT
pragma solidity ^0.8.29;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";

contract MyNFT is ERC721 {
    constructor() ERC721("molten-boat", "TMTLTBOTEN") {}

    function message() public pure returns (string memory) {
        return "Unleash the power of DeFi - ctom4";
    }
}
