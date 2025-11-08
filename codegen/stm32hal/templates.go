package stm32hal

const (
	headerFile = `#ifndef VERA_STM32HAL_H
#ifndef VERA_STM32HAL_H

#include "vera.h"
#include "stm32f2xx_hal_can.h"

vera_err_t vera_decode_stm32hal_rx_frame(
	CAN_RxHeaderTypeDef*    frame,
	uint8_t*                data,
	vera_decoding_result_t* result
);

#endif // VERA_STM32HAL_H`
	sourceFile = `#include "vera_smt32hal.h"
#include <string.h>

vera_err_t vera_decode_stm32hal_rx_frame(
	CAN_RxHeaderTypeDef*    frame,
	uint8_t*                data,
	vera_decoding_result_t* result
) {
	vera_can_rx_frame_t vera_frame = {
		.id             = frame->IDE == CAN_ID_EXT ? frame->ExtId : frame->StdId,
		.dlc            = frame->DLC * 8,
		.is_extended_id = frame->IDE == CAN_ID_EXT ? true : false,
		.timestamp      = frame->Timestamp
	};
	memcpy(vera_frame.data, data, frame->DLC);

	return vera_decode_can_frame(vera_frame, result);
}`
)
