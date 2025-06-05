import { Coordinate, RequestStatus } from "~/apiClient/apiSchemas";
import { RequestId } from "./apiClient/apiParameters";
import { Chair } from "./apiClient/apiSchemas";
type AccessToken = string;

export type ClientAppRequest = {
  status?: RequestStatus;
  payload: Partial<{
    request_id: RequestId;
    coordinate: Partial<{
      pickup: Coordinate;
      destination: Coordinate;
    }>;
    chair?: Chair;
  }>;
};

export type User = {
  id: string;
  name: string;
  accessToken: AccessToken;
  request?: ClientAppRequest;
};
