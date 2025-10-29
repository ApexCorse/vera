#include "vera.h"
#include "unity/unity.h"

void setUp(void) {}
void tearDown(void) {}

void test_successful_decoding(void) {
	vera_can_rx_frame_t frame = {
		.id = 0x7b,
		.dlc = 6,
		.data = {0x42, 0x58, 0x7d, 0xf4, 0x0c, 0xe5},
	};
	vera_decoding_result_t result = {
		.n_signals = 0,
		.decoded_signals = NULL
	};

	vera_err_t err = vera_decode_can_frame(&frame, &result);
	TEST_ASSERT_EQUAL(vera_err_ok, err);
	TEST_ASSERT_NOT_NULL(result.decoded_signals);
	TEST_ASSERT_EQUAL(2, result.n_signals);

	vera_decoded_signal_t* decoded_signals = result.decoded_signals;
	TEST_ASSERT_EQUAL_STRING("RPM", decoded_signals[0].unit);
	TEST_ASSERT_EQUAL_STRING("EngineSpeed", decoded_signals[0].name); 
	TEST_ASSERT_EQUAL_STRING("Engine/Metrics/Speed", decoded_signals[0].topic);
	TEST_ASSERT_FLOAT_WITHIN(0.01, 5.412, decoded_signals[0].value);
	TEST_ASSERT_EQUAL_STRING("ºC", decoded_signals[1].unit);
	TEST_ASSERT_EQUAL_STRING("BatteryTemperature", decoded_signals[1].name);
	TEST_ASSERT_EQUAL_FLOAT(206.3125, decoded_signals[1].value);
}

int main(void) {
	setvbuf(stdout, NULL, _IONBF, 0); // Disable stdout buffering
	UNITY_BEGIN();

	RUN_TEST(test_successful_decoding);
	return UNITY_END();
}

