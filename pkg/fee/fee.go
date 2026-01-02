package fee

import (
	"errors"
	"math"
)

// FeeCalculator calculates transaction fees based on on-chain income
type FeeCalculator struct {
	baseFee     int64   // Base fee in smallest unit
	minFee      int64   // Minimum fee
	maxFee      int64   // Maximum fee
	scaleFactor float64 // How aggressively fees scale with income
}

// NewFeeCalculator creates a new fee calculator with default parameters
func NewFeeCalculator() *FeeCalculator {
	return &FeeCalculator{
		baseFee:     100,  // Base fee of 100 units
		minFee:      10,   // Minimum fee of 10 units
		maxFee:      10000, // Maximum fee of 10000 units
		scaleFactor: 0.5, // 50% scaling factor for more visible differences
	}
}

// NewCustomFeeCalculator creates a fee calculator with custom parameters
func NewCustomFeeCalculator(baseFee, minFee, maxFee int64, scaleFactor float64) *FeeCalculator {
	return &FeeCalculator{
		baseFee:     baseFee,
		minFee:      minFee,
		maxFee:      maxFee,
		scaleFactor: scaleFactor,
	}
}

// CalculateFee calculates the transaction fee based on the sender's on-chain income
// Income-based scaling: lower income = lower fees, higher income = higher fees
func (fc *FeeCalculator) CalculateFee(onChainIncome int64, transactionValue int64) (int64, error) {
	if onChainIncome < 0 {
		return 0, errors.New("on-chain income cannot be negative")
	}

	if transactionValue < 0 {
		return 0, errors.New("transaction value cannot be negative")
	}

	// For users with very low or zero income, use minimum fee
	if onChainIncome < 100 {
		return fc.minFee, nil
	}

	// Calculate fee based on income using logarithmic scaling
	// This ensures fees grow slowly with income, keeping the network accessible
	incomeFactor := math.Log(float64(onChainIncome)+1) / math.Log(10)
	calculatedFee := int64(float64(fc.baseFee) * incomeFactor * fc.scaleFactor)

	// Apply bounds
	if calculatedFee < fc.minFee {
		return fc.minFee, nil
	}
	if calculatedFee > fc.maxFee {
		return fc.maxFee, nil
	}

	return calculatedFee, nil
}

// CalculateFeeWithReputation calculates fee with reputation discount
// Higher reputation = lower fees (as a reward for honest participation)
// Formula: discount = min(reputation / 10000, 0.5)
// This provides up to 50% discount at 5000+ reputation points
func (fc *FeeCalculator) CalculateFeeWithReputation(onChainIncome int64, transactionValue int64, reputation int64) (int64, error) {
	baseFee, err := fc.CalculateFee(onChainIncome, transactionValue)
	if err != nil {
		return 0, err
	}

	// Apply reputation discount (up to 50% off for high reputation)
	reputationDiscount := float64(reputation) / 10000.0 // Max 50% discount at 5000 reputation
	if reputationDiscount > 0.5 {
		reputationDiscount = 0.5
	}

	discountedFee := int64(float64(baseFee) * (1.0 - reputationDiscount))

	// Ensure we don't go below minimum fee
	if discountedFee < fc.minFee {
		return fc.minFee, nil
	}

	return discountedFee, nil
}

// GetBaseFee returns the base fee
func (fc *FeeCalculator) GetBaseFee() int64 {
	return fc.baseFee
}

// GetMinFee returns the minimum fee
func (fc *FeeCalculator) GetMinFee() int64 {
	return fc.minFee
}

// GetMaxFee returns the maximum fee
func (fc *FeeCalculator) GetMaxFee() int64 {
	return fc.maxFee
}

// SetParameters allows updating fee parameters
func (fc *FeeCalculator) SetParameters(baseFee, minFee, maxFee int64, scaleFactor float64) error {
	if minFee > baseFee || baseFee > maxFee {
		return errors.New("invalid fee parameters: minFee <= baseFee <= maxFee")
	}
	if scaleFactor <= 0 {
		return errors.New("scale factor must be positive")
	}

	fc.baseFee = baseFee
	fc.minFee = minFee
	fc.maxFee = maxFee
	fc.scaleFactor = scaleFactor
	return nil
}
