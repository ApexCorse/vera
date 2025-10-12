#include "vera.h"
#include "unity/unity.h"

void setUp(void) {}
void tearDown(void) {}

void test_successful_decoding(void) {
	vera_can_rx_frame_t frame = {
		.id = 0x7b,
		.dlc = 2,
		.data = {0x45, 0xaa},
	};
	vera_decoded_signal_t* decoded_signals = NULL;

	vera_err_t err = vera_decode_can_frame(&frame, decoded_signals);
	TEST_ASSERT_EQUAL(vera_err_ok, err);
}

int main(void) {
	UNITY_BEGIN();

	RUN_TEST(test_successful_decoding);
	return UNITY_END();
}

