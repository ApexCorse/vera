#include "vera.h"
#include "unity/unity.h"

void setUp(void) {}
void tearDown(void) {}

void test_successful_decoding(void) {
	vera_can_rx_frame_t frame = {
		.id = 0x7b,
		.dlc = 4,
		.data = {0x42, 0x58, 0x7d, 0xf4},
	};
	vera_decoded_signal_t* decoded_signals = NULL;

	vera_err_t err = vera_decode_can_frame(&frame, &decoded_signals);
	TEST_ASSERT_EQUAL(vera_err_ok, err);
	TEST_ASSERT_NOT_NULL(decoded_signals);

	TEST_ASSERT_EQUAL_STRING("RPM", decoded_signals[0].unit);
	TEST_ASSERT_EQUAL_STRING("EngineSpeed", decoded_signals[0].name); 
	TEST_ASSERT_FLOAT_WITHIN(0.01, 5.412, decoded_signals[0].value);
}

int main(void) {
	UNITY_BEGIN();

	RUN_TEST(test_successful_decoding);
	return UNITY_END();
}

