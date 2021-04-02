package wallet

const (
	P2PK = iota
	P2PKH
	P2MS
	P2SH_P2WPKH
	P2SH_P2WSH
	P2WPKH
	P2WSH
)

// EstimateTxSize makes an estimation of the virtual size of a transaction for
// which is required to specify the type of the inputs and outputs according to
// those of the Bitcoin standard (P2PK, P2PKH, P2MS, P2SH(P2WPKH), P2SH(P2WSH),
// P2WPKH, P2WSH).
// In case some inputs or outputs are of type P2MS, it is mandatory to pass
// their redeem script sizes as auxiliary slices in accordance.
func EstimateTxSize(
	inScriptTypes, inAuxiliaryRedeemScriptSize, inAuxiliaryWitnessSize,
	outScriptTypes, outAuxiliaryRedeemScriptSize []int,
) int {
	baseSize := calcTxSize(
		false,
		inScriptTypes, inAuxiliaryRedeemScriptSize, inAuxiliaryWitnessSize,
		outScriptTypes, outAuxiliaryRedeemScriptSize,
	)
	totalSize := calcTxSize(
		true,
		inScriptTypes, inAuxiliaryRedeemScriptSize, inAuxiliaryWitnessSize,
		outScriptTypes, outAuxiliaryRedeemScriptSize,
	)

	weight := baseSize*3 + totalSize
	vsize := (weight + 3) / 4

	return vsize
}

func calcTxSize(
	withWitness bool,
	inScriptTypes, inAuxiliaryRedeemScriptSize, inAuxiliaryWitnessSize,
	outScriptTypes, outAuxiliaryRedeemScriptSize []int,
) int {
	txSize := calcTxBaseSize(
		inScriptTypes, inAuxiliaryRedeemScriptSize,
		outScriptTypes, outAuxiliaryRedeemScriptSize,
	)
	if withWitness {
		txSize += calcTxWitnessSize(
			inScriptTypes, inAuxiliaryWitnessSize,
			outScriptTypes,
		)
	}
	return txSize
}

var (
	scripsigtSizeByScriptType = map[int]int{
		P2PK:        140, // len + opcode + sig + opcode + pubkey uncompressed
		P2PKH:       108, // len + opcode + sig + opcode + pubkey
		P2SH_P2WPKH: 23,  // len + p2wpkh script
		P2SH_P2WSH:  35,  // len + p2wsh script
		P2WPKH:      1,   // no scriptsig, still len is serialized
		P2WSH:       1,   // no scriptsig
	}
	scriptPubKeySizeByScriptType = map[int]int{
		P2PK:        67, // len + pubkey uncompressed + opcode
		P2PKH:       26, // len + opcodes (3) + hash(pubkey) + opcodes (2)
		P2SH_P2WPKH: 24, // len + opcodes (2) + hash(script) + opcode
		P2SH_P2WSH:  24, // len + opcodes (2) + hash(script) + opcode
		P2WPKH:      23, // len + opcodes (2) + hash(script)
		P2WSH:       35, // len + opcodes (2) + hash(script)
	}
)

func calcTxBaseSize(
	inScriptTypes, inAuxiliaryRedeemScriptSize,
	outScriptTypes, outAuxiliaryRedeemScriptSize []int,
) int {
	// hash + index + sequence
	inBaseSize := 40
	insSize := 0
	auxCount := 0
	for _, scriptType := range inScriptTypes {
		scriptSize, ok := scripsigtSizeByScriptType[scriptType]
		if !ok {
			scriptSize = inAuxiliaryRedeemScriptSize[auxCount]
			auxCount++
		}
		insSize += inBaseSize + scriptSize
	}

	// asset + value + nonce commitments
	outBaseSize := 33 + 33 + 33
	outsSize := 0
	auxCount = 0
	for _, scriptType := range outScriptTypes {
		scriptSize, ok := scriptPubKeySizeByScriptType[scriptType]
		if !ok {
			scriptSize = outAuxiliaryRedeemScriptSize[auxCount]
			auxCount++
		}
		outsSize += outBaseSize + scriptSize
	}
	// size of unconf fee out
	// asset + unconf value + empty script + empty nonce
	outsSize += 33 + 9 + 1 + 1

	return 9 +
		varIntSerializeSize(uint64(len(inScriptTypes))) +
		varIntSerializeSize(uint64(len(outScriptTypes)+1)) +
		insSize + outsSize
}

func calcTxWitnessSize(
	inScriptTypes, inAuxiliaryWitnessSize,
	outScriptTypes []int,
) int {
	insSize := 0
	auxCount := 0
	for _, scriptType := range inScriptTypes {
		if scriptType == P2SH_P2WPKH || scriptType == P2WPKH {
			// len + witness[sig,pubkey] + no issuance proof + no token proof + no pegin
			insSize += (1 + 107 + 1 + 1 + 1)
		}
		if scriptType == P2SH_P2WSH || scriptType == P2WSH {
			insSize += inAuxiliaryWitnessSize[auxCount]
			auxCount++
		}
	}

	numOutputs := len(outScriptTypes)
	// size(range proof) + proof + size(surjection proof) + proof
	outsSize := (3 + 4174 + 1 + 131) * numOutputs
	// size of proofs for unconf fee out
	outsSize += 1 + 1

	return insSize + outsSize
}
