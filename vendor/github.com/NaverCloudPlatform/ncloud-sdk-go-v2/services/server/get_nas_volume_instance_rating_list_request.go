/*
 * server
 *
 * <br/>https://ncloud.apigw.ntruss.com/server/v2
 *
 * API version: 2018-10-18T06:16:13Z
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package server

type GetNasVolumeInstanceRatingListRequest struct {

	// 측정종료시간
EndTime *string `json:"endTime"`

	// 측정간격
Interval *string `json:"interval"`

	// NAS볼륨인스턴스번호
NasVolumeInstanceNo *string `json:"nasVolumeInstanceNo"`

	// 측정시작시간
StartTime *string `json:"startTime"`
}
