package codegen

const (
	includeFile = `#ifndef VERA_H
#define VERA_H

#include <stdbool.h>
#include <stdlib.h>

#define CAN_MAX_DATA_LEN 8

typedef struct {
	uint32_t id;
	uint8_t  dlc;
	uint8_t  data[CAN_MAX_DATA_LEN];

	bool is_extended_id;
	bool is_rtr;
	bool is_fd;
	bool bit_rate_switch;
	bool error_state_indicator;

	uint64_t timestamp;
} vera_can_rx_frame_t;

typedef struct {
	char*   name;
	uint8_t start_bit;
	uint8_t dlc;
	uint8_t endianness;
	bool    sign;
	uint8_t integer_figures;
	uint8_t decimal_figures;
	float   factor;
	float   offset;
	float   min;
	float   max;
	char*   unit;
	char**  receivers;
	char*   topic;
} vera_signal_t;

typedef struct {
	uint32_t       id;
	char*          name;
	uint8_t        dlc;
	char*          transmitter;
	vera_signal_t* signals;
	uint8_t        n_signals;
} vera_message_t;

typedef struct {
	char*    name;
	char*    unit;
	float    value;
	uint64_t timestamp;
	char*    topic;
} vera_decoded_signal_t;

typedef struct {
	uint8_t n_signals;
	vera_decoded_signal_t* decoded_signals;
} vera_decoding_result_t;

typedef enum {
	vera_err_ok,
	vera_err_allocation,
	vera_err_out_of_bounds
} vera_err_t;

vera_err_t vera_decode_can_frame(
	vera_can_rx_frame_t*   frame,
	vera_decoding_result_t* result
);

#endif // VERA_H`
	sourceFileIncludes = `#include <strings.h>
#include <stdio.h>
#include <math.h>`
	decodeMessageFunc = `vera_err_t _decode_message(
	vera_can_rx_frame_t*    frame,
	vera_message_t*         message,
	vera_decoding_result_t* result
) {
	result->decoded_signals = (vera_decoded_signal_t*)malloc(sizeof(vera_decoded_signal_t)*message->n_signals);
	if (!result->decoded_signals) return vera_err_allocation;

	for (uint8_t i = 0; i < message->n_signals; i++) {
		vera_err_t err = _decode_signal(
			frame,
			message->signals + i,
			result->decoded_signals + i
		);
		if (err != vera_err_ok) {
			free(result->decoded_signals);
			result->decoded_signals = NULL;
			return err;
		}
		result->n_signals++;
	}

	return vera_err_ok;
}`
	decodeSignalFunc = `vera_err_t _decode_signal(
	vera_can_rx_frame_t*   frame,
	vera_signal_t*         signal,
	vera_decoded_signal_t* res
) {
	res->name = strdup(signal->name);
	if (!res->name) return vera_err_allocation;

	res->unit = strdup(signal->unit);
	if (!res->unit) {
		free(res->name);
		return vera_err_allocation;
	}

	if (strlen(signal->topic) > 0) {
		res->topic = strdup(signal->topic);
		if (!res->topic) {
			free(res->name);
			free(res->unit);
			return vera_err_allocation;
		}
	}

	if (signal->start_bit >= frame->dlc || signal->start_bit + signal->dlc > frame->dlc) {
		free(res->name);
		free(res->unit);
		return vera_err_out_of_bounds;		
	}

	uint64_t data = _get_payload_by_start_and_length(
		frame->data,
		signal->start_bit,
		signal->dlc
	);

	if (signal->integer_figures || signal->decimal_figures)
		res->value = _parse_fixed_point_float(
			data,
			signal->integer_figures,
			signal->decimal_figures
		);
	else res->value = _parse_float_directly(data);


	res->value *= signal->factor;
	res->value += signal->offset;
	if (res->value < signal->min)
		res->value = signal->min;
	if (res->value > signal->max)
		res->value = signal->max;

	return vera_err_ok;
}`
	utilFunctions = `// Needs previous validation
float _parse_fixed_point_float(
	uint32_t value,
	uint8_t  integer_figures,
	uint8_t  decimal_figures
) {
	float parsed_value = 0.0;

	for (int i = 0; i < decimal_figures; i++) {
		parsed_value += ((value >> i) & 1) * pow(2, (float)(i - decimal_figures));
	}

	for (int i = 0; i < integer_figures; i++) {
		parsed_value += ((value >> (decimal_figures + i)) & 1) * pow(2, (float)i);
	}

	return parsed_value;
}

float _parse_float_directly(uint32_t value) {
	typedef union {
		uint32_t u;
		float f;
	} convert_union;
	convert_union cu;
	cu.u = value;

	return cu.f;
}

uint64_t _get_payload_by_start_and_length(uint8_t* payload, uint8_t start, uint8_t length) {
	uint64_t res = 0LLU;

	for (uint8_t i = start; i < start + length; i++) {
		uint8_t payload_index = i / 8;
		uint8_t byte = payload[payload_index];
		uint8_t shift_right = 7 - (i - start - payload_index * 8);
		uint8_t shift_left = length + start - 1 - i;
	
		res |= ((payload[payload_index] >> shift_right) & 1) << shift_left;
	}

	return res;
}
`
)
