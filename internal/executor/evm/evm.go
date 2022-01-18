package evm

import (
	"github.com/ethereum/evmc/v10/bindings/go/evmc"
)

// Run implements the runtime interface
func Run(recipient evmc.Address, sender evmc.Address, value evmc.Hash, input []byte, gas int64, static bool, codeAddress evmc.Address, host evmc.HostContext, rev evmc.Revision) ([]byte, int64, evmc.Address, error) {

	contract := acquireState()
	contract.resetReturnData()

	contract.recipient = recipient
	contract.sender = sender
	contract.value = value
	contract.input = input
	contract.static = static
	contract.code = host.GetCode(codeAddress)
	contract.gas = uint64(gas)
	contract.host = host
	contract.rev = rev

	contract.bitmap.setCode(contract.code)

	ret, err := contract.Run()

	// We are probably doing this append magic to make sure that the slice doesn't have more capacity than it needs
	var returnValue []byte
	returnValue = append(returnValue[:0], ret...)

	gasLeft := contract.gas

	releaseState(contract)

	if err != nil && err != errRevert {
		gasLeft = 0
	}
	return returnValue, int64(gasLeft), evmc.Address{}, err
}
