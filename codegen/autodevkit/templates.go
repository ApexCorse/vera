package autodevkit

const (
	headerFile = `#ifndef VERA_AUTODEVKIT_H
#define VERA_AUTODEVKIT_H

#include "vera.h"
#include "can_lld.h"

vera_err_t vera_decode_autodevkit_rx_frame(CANRxFrame* frame, vera_decoding_result_t* result);

%s

#endif // VERA_AUTODEVKIT_H`
	sourceFile = `#include "vera_autodevkit.h"
#include <string.h>

vera_err_t vera_decode_autodevkit_rx_frame(CANRxFrame* frame, vera_decoding_result_t* result) {
	vera_can_rx_frame_t vera_frame = {
		.id             = frame->ID,
		.dlc            = frame->DLC,
		.is_extended_id = frame->TYPE,
		.is_fd          = frame->OPERATION == 0x01U ? true : false
	};
	memcpy(vera_frame.data, frame->data8, frame->DLC);

	return vera_decode_can_frame(&vera_frame, result);
}

%s`
)
