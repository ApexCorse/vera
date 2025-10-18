package codegen

const (
	includeFile = `#include <stdbool.h>
#include <stdlib.h>

#define CAN_FD_MAX_DATA_LEN 64

typedef struct {
	uint32_t 	id;
	uint8_t 	dlc;
	uint8_t		data[CAN_FD_MAX_DATA_LEN];

	bool	is_extended_id;
	bool	is_rtr;
	bool	is_fd;
	bool	bit_rate_switch;
	bool	error_state_indicator;

	uint64_t	timestamp;
} vera_can_rx_frame_t;

typedef struct {
	char*			name;
	uint8_t		start_byte;
	uint8_t		dlc;
	uint8_t		endianness;
	bool			sign;
	float			factor;
	float			offset;
	float			min;
	float			max;
	char*			unit;
	char**		receivers;
} vera_signal_t;

typedef struct {
	uint32_t				id;
	char*						name;
	uint8_t					dlc;
	char*						transmitter;
	vera_signal_t* 	signals;
	uint8_t					n_signals;
} vera_message_t;

typedef struct {
	char*			name;
	char*			unit;
	float			value;
	uint64_t 	timestamp;
} vera_decoded_signal_t;

typedef enum {
	vera_err_ok,
	vera_err_allocation,
	vera_err_out_of_bounds
} vera_err_t;

vera_err_t vera_decode_can_frame(
	vera_can_rx_frame_t* 			frame,
	vera_decoded_signal_t** 	decoded_signals
);`
	sourceFileIncludes = `#include <strings.h>
#include <stdio.h>`
	decodeMessageFunc = `vera_err_t _decode_message(
	vera_can_rx_frame_t* 			frame,
	vera_message_t* 					message,
	vera_decoded_signal_t** 	decoded_signals
) {
	*decoded_signals = (vera_decoded_signal_t*)malloc(sizeof(vera_decoded_signal_t)*message->n_signals);
	if (!*decoded_signals) return vera_err_allocation;

	for (uint8_t i = 0; i < message->n_signals; i++) {
		vera_err_t err = _decode_signal(
			frame,
			&(message->signals[i]),
			(decoded_signals[i])
		);
		if (err != vera_err_ok) {
			free(decoded_signals);
			return err;
		}
	}

	return vera_err_ok;
}`
	decodeSignalFunc = `vera_err_t _decode_signal(
	vera_can_rx_frame_t* 		frame,
	vera_signal_t*					signal,
	vera_decoded_signal_t* 	res
) {
	res->name = strdup(signal->name);
	if (!res->name) return vera_err_allocation;

	res->unit = strdup(signal->unit);
	if (!res->unit) {
		free(res->name);
		return vera_err_allocation;
	}
	
	if (signal->start_byte >= frame->dlc || signal->start_byte + signal->dlc > frame->dlc) {
		free(res->name);
		free(res->unit);
		return vera_err_out_of_bounds;		
	}

	uint8_t payload[signal->dlc];
	memcpy(payload, frame->data + signal->start_byte, signal->dlc);

	uint32_t data = 0;
	for (uint8_t i = 0; i < signal->dlc; i++) {
		if (signal->endianness == 0) {
			data |= ((uint32_t)payload[i] << i * 8);
		} else {
			data |= ((uint32_t)payload[i] << (signal->dlc-1-i) * 8);
		}
	}

	typedef union {
		uint32_t u;
		float f;
	} convert_union;
	convert_union cu;
	cu.u = data;
	res->value = cu.f;

	return vera_err_ok;
}`
)
