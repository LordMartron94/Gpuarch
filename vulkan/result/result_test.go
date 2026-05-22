package result

import (
	"errors"
	"testing"

	"gpuarch/vulkan/bindings"
)

func TestVulkanResultIsFailure(t *testing.T) {
	if VulkanResultIsFailure(bindings.VK_SUCCESS) {
		t.Fatal("VK_SUCCESS is not a failure")
	}
	if VulkanResultIsFailure(bindings.VK_NOT_READY) {
		t.Fatal("VK_NOT_READY is not a failure")
	}
	if !VulkanResultIsFailure(bindings.VK_ERROR_DEVICE_LOST) {
		t.Fatal("VK_ERROR_DEVICE_LOST is a failure")
	}
}

func TestVulkanResultToError(t *testing.T) {
	if err := VulkanResultToError(bindings.VK_SUCCESS); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := VulkanResultToError(bindings.VK_NOT_READY); err == nil {
		t.Fatal("expected error for VK_NOT_READY")
	} else {
		var resultErr *VulkanResultError
		if !errors.As(err, &resultErr) || resultErr.Code != bindings.VK_NOT_READY {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestVulkanResultToErrorIfFailure(t *testing.T) {
	if err := VulkanResultToErrorIfFailure(bindings.VK_INCOMPLETE); err != nil {
		t.Fatalf("expected nil for VK_INCOMPLETE, got %v", err)
	}
	if err := VulkanResultToErrorIfFailure(bindings.VK_ERROR_OUT_OF_HOST_MEMORY); err == nil {
		t.Fatal("expected error for failure code")
	}
}

func TestVulkanResultAssertSuccessOr(t *testing.T) {
	if err := VulkanResultAssertSuccessOr(bindings.VK_INCOMPLETE, bindings.VK_INCOMPLETE); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := VulkanResultAssertSuccessOr(bindings.VK_TIMEOUT, bindings.VK_INCOMPLETE); err == nil {
		t.Fatal("expected error when code not allowed")
	}
}

func TestVulkanResultName(t *testing.T) {
	if VulkanResultName(bindings.VK_SUCCESS) != "VK_SUCCESS" {
		t.Fatal("expected VK_SUCCESS name")
	}
	if VulkanResultName(bindings.VkResult(424242)) != "VkResult(424242)" {
		t.Fatal("expected formatted unknown name")
	}
}
