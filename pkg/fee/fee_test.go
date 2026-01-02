package fee

import (
	"testing"
)

func TestFeeCalculator_BasicFee(t *testing.T) {
	fc := NewFeeCalculator()
	
	// Test with zero income (should return minimum fee)
	fee, err := fc.CalculateFee(0, 1000)
	if err != nil {
		t.Fatalf("Failed to calculate fee: %v", err)
	}
	
	if fee != fc.GetMinFee() {
		t.Errorf("Expected minimum fee %d for zero income, got %d", fc.GetMinFee(), fee)
	}
}

func TestFeeCalculator_IncomeScaling(t *testing.T) {
	fc := NewFeeCalculator()
	
	// Test that fees increase with income
	fee1, _ := fc.CalculateFee(100, 1000)
	fee2, _ := fc.CalculateFee(1000, 1000)
	fee3, _ := fc.CalculateFee(10000, 1000)
	
	if fee1 >= fee2 {
		t.Errorf("Expected fee to increase with income: %d >= %d", fee1, fee2)
	}
	
	if fee2 >= fee3 {
		t.Errorf("Expected fee to increase with income: %d >= %d", fee2, fee3)
	}
}

func TestFeeCalculator_ReputationDiscount(t *testing.T) {
	fc := NewFeeCalculator()
	
	baseFee, _ := fc.CalculateFee(10000, 1000)
	discountedFee, _ := fc.CalculateFeeWithReputation(10000, 1000, 1000)
	
	if discountedFee >= baseFee {
		t.Errorf("Expected discounted fee %d to be less than base fee %d", discountedFee, baseFee)
	}
}

func TestFeeCalculator_MaxDiscount(t *testing.T) {
	fc := NewFeeCalculator()
	
	// Test with very high reputation (should cap at 50% discount)
	baseFee, _ := fc.CalculateFee(10000, 1000)
	discountedFee, _ := fc.CalculateFeeWithReputation(10000, 1000, 100000)
	
	// Should be at least 50% of base fee
	minExpectedFee := baseFee / 2
	if discountedFee < minExpectedFee {
		t.Errorf("Discount too high: fee %d is less than 50%% of base %d", discountedFee, baseFee)
	}
}

func TestFeeCalculator_MinFeeFloor(t *testing.T) {
	fc := NewFeeCalculator()
	
	// Even with high reputation and low income, fee should not go below minimum
	fee, _ := fc.CalculateFeeWithReputation(100, 1000, 10000)
	
	if fee < fc.GetMinFee() {
		t.Errorf("Fee %d is below minimum %d", fee, fc.GetMinFee())
	}
}

func TestFeeCalculator_MaxFee(t *testing.T) {
	fc := NewFeeCalculator()
	
	// Test with extremely high income
	fee, _ := fc.CalculateFee(999999999, 1000)
	
	if fee > fc.GetMaxFee() {
		t.Errorf("Fee %d exceeds maximum %d", fee, fc.GetMaxFee())
	}
}

func TestFeeCalculator_NegativeValues(t *testing.T) {
	fc := NewFeeCalculator()
	
	_, err := fc.CalculateFee(-100, 1000)
	if err == nil {
		t.Error("Expected error for negative income")
	}
	
	_, err = fc.CalculateFee(1000, -100)
	if err == nil {
		t.Error("Expected error for negative transaction value")
	}
}

func TestFeeCalculator_CustomParameters(t *testing.T) {
	fc := NewCustomFeeCalculator(200, 20, 20000, 0.02)
	
	if fc.GetBaseFee() != 200 {
		t.Errorf("Expected base fee 200, got %d", fc.GetBaseFee())
	}
	
	if fc.GetMinFee() != 20 {
		t.Errorf("Expected min fee 20, got %d", fc.GetMinFee())
	}
	
	if fc.GetMaxFee() != 20000 {
		t.Errorf("Expected max fee 20000, got %d", fc.GetMaxFee())
	}
}

func TestFeeCalculator_SetParameters(t *testing.T) {
	fc := NewFeeCalculator()
	
	err := fc.SetParameters(150, 15, 15000, 0.015)
	if err != nil {
		t.Fatalf("Failed to set parameters: %v", err)
	}
	
	if fc.GetBaseFee() != 150 {
		t.Errorf("Expected base fee 150, got %d", fc.GetBaseFee())
	}
	
	// Test invalid parameters
	err = fc.SetParameters(10, 100, 1000, 0.01) // min > base
	if err == nil {
		t.Error("Expected error for invalid parameters")
	}
	
	err = fc.SetParameters(100, 10, 50, 0.01) // base > max
	if err == nil {
		t.Error("Expected error for invalid parameters")
	}
	
	err = fc.SetParameters(100, 10, 1000, -0.01) // negative scale factor
	if err == nil {
		t.Error("Expected error for negative scale factor")
	}
}
