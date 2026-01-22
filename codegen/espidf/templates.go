package espidf

const (
	headerFile = `#ifndef VERA_ESPIDF_H
#define VERA_ESPIDF_H

#include "vera.h"
#include "driver/twai.h"

vera_err_t vera_decode_espidf_rx_frame(const twai_frame_t* frame, vera_decoding_result_t* result);

%s

#endif // VERA_ESPIDF_H`

	sourceFile = `#include "vera_espidf.h"
#include <string.h>

vera_err_t vera_decode_espidf_rx_frame(const twai_frame_t* frame, vera_decoding_result_t* result) {
    vera_can_rx_frame_t vera_frame = {
        .id = frame->header.id,
        .dlc = frame->header.dlc,
        .is_extended_id = frame->header.ide,
        .is_rtr = frame->header.rtr,
        .is_fd = frame->header.fdf,
        .bit_rate_switch = frame->header.brs,
        .error_state_indicator = frame->header.esi
    };
    memcpy(vera_frame.data, frame->buffer, frame->header.dlc > 8 ? 8 : frame->header.dlc);

    return vera_decode_can_frame(&vera_frame, result);
}

%s`
)
