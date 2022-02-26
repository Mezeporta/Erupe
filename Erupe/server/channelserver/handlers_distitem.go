package channelserver

import (
	"encoding/hex"
	// "io/ioutil"
	// "path/filepath"

	"github.com/Solenataris/Erupe/network/mhfpacket"
)

func handleMsgMhfEnumerateDistItem(s *Session, p mhfpacket.MHFPacket) {
    pkt := p.(*mhfpacket.MsgMhfEnumerateDistItem)
    // uint16 number of entries
    // 446 entry block
    // uint32 claimID
    // 00 00 00 00 00 00
    // uint16 timesClaimable
    // 00 00 00 00 FF FF FF FF FF FF FF FF FF FF FF FF 00 00 00 00 00 00 00 00 00
    // uint8 stringLength
    // string nullTermString
    data, _ := hex.DecodeString("0001000000010000000000000000002000000000FFFFFFFFFFFFFFFFFFFFFFFF0000000000000000002F323020426F7820457870616E73696F6E73000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
    doAckBufSucceed(s, pkt.AckHandle, data)

}

func handleMsgMhfApplyDistItem(s *Session, p mhfpacket.MHFPacket) {
    // 0052a49100011f00000000000000010274db99 equipment box page
    // 0052a48f00011e0000000000000001195dda5c item box page
    // 0052a49400010700003ae30000000132d3a4d6 Item ID 3AE3
    // HEADER:
    // int32: Unique item hash for tracking server side purchases? Swapping across items didn't change image/cost/function etc.
    // int16: Number of distributed item types
    // ITEM ENTRY
    // int8:  distribution type
    // 00 = legs, 01 = Head, 02 = Chest, 03 = Arms, 04 = Waist, 05 = Melee, 06 = Ranged, 07 = Item, 08 == furniture
    // ids are wrong shop displays in random order
    // 09 = Nothing, 10 = Null Point, 11 = Festi Point, 12 = Zeny, 13 = Null, 14 = Null Points, 15 = My Tore points
    // 16 = Restyle Point, 17 = N Points, 18 = Nothing, 19 = Gacha Coins, 20 = Trial Gacha Coins, 21 = Frontier points
    // 22 = ?, 23 = Guild Points, 30 = Item Box Page, 31 = Equipment Box Page
    // int16: Unk
    // int16: Item when type 07
    // int16: Unk
    // int16: Number delivered in batch
    // int32: Unique item hash for tracking server side purchases? Swapping across items didn't change image/cost/function etc.
    pkt := p.(*mhfpacket.MsgMhfApplyDistItem)
    if pkt.RequestType == 0 {
        doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
    } else if pkt.RequestType == 0x00000001 {
        data, _ := hex.DecodeString("0052a494001e0f000000010000000132d3a4d60f000000020000000132d3a4d60f000000030000000132d3a4d60f000000040000000132d3a4d60f000000050000000132d3a4d60f000000060000000132d3a4d60f000000070000000132d3a4d60f000000080000000132d3a4d60f000000090000000132d3a4d60f0000000a0000000132d3a4d60f0000000b0000000132d3a4d60f0000000c0000000132d3a4d60f0000000d0000000132d3a4d60f0000000e0000000132d3a4d60f0000000f0000000132d3a4d60f000000100000000132d3a4d60f000000110000000132d3a4d60f000000120000000132d3a4d60f000000130000000132d3a4d60f000000140000000132d3a4d60f000000150000000132d3a4d60f000000160000000132d3a4d60f000000170000000132d3a4d60f000000180000000132d3a4d60f000000190000000132d3a4d60f0000001a0000000132d3a4d60f0000001b0000000132d3a4d60f0000001c0000000132d3a4d60f0000001d0000000132d3a4d60f0000001e0000000132d3a4d6")
        doAckBufSucceed(s, pkt.AckHandle, data)
    } else {
        doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
    }
}

func handleMsgMhfAcquireDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireDistItem)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetDistDescription(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetDistDescription)
	// string for the associated message
	data, _ := hex.DecodeString("007E43303547656E65726963204974656D20436C61696D204D6573736167657E4330300D0A596F752067657420736F6D65206B696E64206F66206974656D732070726F6261626C792E00000100")
	//data, _ := hex.DecodeString("0075b750c1c2b17ac1cab652a1757e433035b8cbb3c6bd63c258b169aa41b0c87e433030a1760a0aa175b8cbb3c6bd63c258b169aa41b0c8a176a843c1caa44a31a6b8a141a569c258b169a2b0add30aa8a4a6e2aabaa175b8cbb3c6bd63a176a2b0adb6a143b3cca668a569c258b169a2b4adb6a14300000100")
	doAckBufSucceed(s, pkt.AckHandle, data)
}
