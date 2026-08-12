#include "vera.h"
#include "unity/unity.h"

void setUp(void) {}
void tearDown(void) {}

void test_successful_decoding(void) {
	vera_can_rx_frame_t frame = {
		.id = 0x7b,
		.dlc = 8,
		.data = {0x00, 0x00, 0x7d, 0xf4, 0x0c, 0xe5, 0x64, 0x10},
	};
	vera_decoded_signal_t signals[vera_n_signals_Message1];
	vera_decoding_result_t result = VERA_DECODING_RESULT(signals);

	vera_err_t err = vera_decode_can_frame(&frame, &result);
	TEST_ASSERT_EQUAL(vera_err_ok, err);
	TEST_ASSERT_NOT_NULL(result.decoded_signals);
	TEST_ASSERT_EQUAL(2, result.n_signals);

	vera_decoded_signal_t* decoded_signals = result.decoded_signals;
	TEST_ASSERT_EQUAL_STRING("RPM", decoded_signals[0].unit);
	TEST_ASSERT_EQUAL_STRING("EngineSpeed", decoded_signals[0].name); 
	TEST_ASSERT_EQUAL_STRING("Engine/Metrics/Speed", decoded_signals[0].topic);
	TEST_ASSERT_FLOAT_WITHIN(0.01, 3224.4, decoded_signals[0].value);
	TEST_ASSERT_EQUAL_STRING("ºC", decoded_signals[1].unit);
	TEST_ASSERT_EQUAL_STRING("BatteryTemperature", decoded_signals[1].name);
	TEST_ASSERT_EQUAL_FLOAT(606, decoded_signals[1].value);
}

void test_successful_encoding(void) {
	vera_can_tx_frame_t frame = {
		.data = {0}
	};
	vera_err_t err = vera_encode_Message1(&frame, 1006985169, 325);

	TEST_ASSERT_EQUAL(vera_err_ok, err);
	TEST_ASSERT_EQUAL(123, frame.id);
	TEST_ASSERT_EQUAL(6, frame.dlc);
	TEST_ASSERT_EQUAL(0x3c, frame.data[0]);
	TEST_ASSERT_EQUAL(0x05, frame.data[1]);
	TEST_ASSERT_EQUAL(0x5f, frame.data[2]);
	TEST_ASSERT_EQUAL(0xd1, frame.data[3]);
	TEST_ASSERT_EQUAL(0x14, frame.data[4]);
	TEST_ASSERT_EQUAL(0x50, frame.data[5]);
	TEST_ASSERT_EQUAL(0x00, frame.data[6]);
	TEST_ASSERT_EQUAL(0x00, frame.data[7]);
}

void test_successful_decoding_clamping(void) {
	vera_can_rx_frame_t frame = {
		.id = 0x7b, // 123
		.dlc = 8,
		.data = {0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00}, // RPM value above max level
	};
	vera_decoded_signal_t signals[vera_n_signals_Message1];
	vera_decoding_result_t result = VERA_DECODING_RESULT(signals);

	vera_err_t err = vera_decode_can_frame(&frame, &result);
	TEST_ASSERT_EQUAL(vera_err_ok, err);
	
	// the program will cut the signal at its max allowed value (8000 RPM)
	TEST_ASSERT_EQUAL_FLOAT(8000.0, result.decoded_signals[0].value); 
}

void test_failing_decoding_unknown_id(void) {
	vera_can_rx_frame_t frame = { .id = 999, .dlc = 8, .data = {0} }; // unknown id
	vera_decoded_signal_t signals[2];
	vera_decoding_result_t result = VERA_DECODING_RESULT(signals);
	result.n_signals = 1;

	vera_err_t err = vera_decode_can_frame(&frame, &result);

	TEST_ASSERT_EQUAL(vera_err_unknown_frame, err);
	TEST_ASSERT_EQUAL(0, result.n_signals);
}

void test_frame_signal_count(void) {
	uint8_t count = 0;
	TEST_ASSERT_EQUAL(vera_err_ok, vera_frame_signal_count(123, &count));
	TEST_ASSERT_EQUAL(2, count);
	TEST_ASSERT_EQUAL(vera_err_ok, vera_frame_signal_count(456, &count));
	TEST_ASSERT_EQUAL(0, count);
	TEST_ASSERT_EQUAL(vera_err_unknown_frame, vera_frame_signal_count(999, &count));
	TEST_ASSERT_EQUAL(0, count);
	TEST_ASSERT_EQUAL(vera_err_null_arg, vera_frame_signal_count(123, NULL));
}

void test_successful_decoding_zero_signal_message(void) {
	vera_can_rx_frame_t frame = { .id = 456, .dlc = 0 };
	vera_decoding_result_t result = {
		.n_signals = 1,
		.decoded_capacity = 0,
		.decoded_signals = NULL
	};

	vera_err_t err = vera_decode_can_frame(&frame, &result);
	TEST_ASSERT_EQUAL(vera_err_ok, err);
	TEST_ASSERT_EQUAL(0, result.n_signals);
}

void test_failing_decoding_insufficient_capacity(void) {
	vera_can_rx_frame_t frame = { .id = 123, .dlc = 8, .data = {0} };
	vera_decoded_signal_t signals[1] = { { .name = "unchanged" } };
	vera_decoding_result_t result = VERA_DECODING_RESULT(signals);

	vera_err_t err = vera_decode_can_frame(&frame, &result);
	TEST_ASSERT_EQUAL(vera_err_insufficient_capacity, err);
	TEST_ASSERT_EQUAL(0, result.n_signals);
	TEST_ASSERT_EQUAL_STRING("unchanged", signals[0].name);
}

void test_failing_decoding_null_arg(void) {
	vera_can_rx_frame_t frame = {
		.id = 123,
		.dlc = 8,
		.data = {0}
	};
	vera_decoding_result_t result = {
		.n_signals = 0,
		.decoded_capacity = 2,
		.decoded_signals = NULL // NULL pointer, it will fail the test
	};

	vera_err_t err = vera_decode_can_frame(&frame, &result);
	TEST_ASSERT_EQUAL(vera_err_null_arg, err);
}

void test_failing_encoding_null_arg(void) {
	vera_err_t err = vera_encode_Message1(NULL, 3000, 400); // NULL argument, it will fail the test
	TEST_ASSERT_EQUAL(vera_err_null_arg, err);
}

void test_failing_decoding_null_frame(void) {
	vera_decoded_signal_t signals[2];
	vera_decoding_result_t result = VERA_DECODING_RESULT(signals);

	vera_err_t err = vera_decode_can_frame(NULL, &result); // NULL frame, it will fail the test
	TEST_ASSERT_EQUAL(vera_err_null_arg, err);
}

void test_failing_decoding_null_result(void) {
	vera_can_rx_frame_t frame = {
		.id = 123,
		.dlc = 8,
		.data = {0}
	};

	vera_err_t err = vera_decode_can_frame(&frame, NULL); // NULL result, it will fail the test
	TEST_ASSERT_EQUAL(vera_err_null_arg, err);
}

void test_failing_decoding_out_of_bounds(void) {
	vera_can_rx_frame_t frame = {
		.id = 123,
		.dlc = 2, // 2 bytes instead of 6, it will fail the test
		.data = {0}
	};
	vera_decoded_signal_t signals[VERA_MAX_SIGNALS_PER_FRAME] = { { .name = "unchanged" } };
	vera_decoding_result_t result = VERA_DECODING_RESULT(signals);

	vera_err_t err = vera_decode_can_frame(&frame, &result);
	TEST_ASSERT_EQUAL(vera_err_out_of_bounds, err);
	TEST_ASSERT_EQUAL(0, result.n_signals);
	TEST_ASSERT_EQUAL_STRING("unchanged", signals[0].name);
}

int main(void) {
	setvbuf(stdout, NULL, _IONBF, 0); // Disable stdout buffering
	UNITY_BEGIN();

	RUN_TEST(test_successful_decoding);
	RUN_TEST(test_successful_encoding);
	RUN_TEST(test_successful_decoding_clamping);
	RUN_TEST(test_failing_decoding_unknown_id);
	RUN_TEST(test_frame_signal_count);
	RUN_TEST(test_successful_decoding_zero_signal_message);
	RUN_TEST(test_failing_decoding_insufficient_capacity);
	RUN_TEST(test_failing_decoding_null_arg);
	RUN_TEST(test_failing_encoding_null_arg);
	RUN_TEST(test_failing_decoding_null_frame);
	RUN_TEST(test_failing_decoding_null_result);
	RUN_TEST(test_failing_decoding_out_of_bounds);

	return UNITY_END();
}
