import { apiRequestWithAuth } from './client';
import { CommandAck, CommandType } from './types';

export interface SendCommandResponse {
  ack: CommandAck;
}

export const sendCommand = async (
  deviceId: string,
  type: CommandType,
  accessToken: string,
  params?: Record<string, unknown>,
  timeoutMs?: number,
  refreshToken?: () => Promise<string>,
  onUnauthorized?: () => Promise<void>
) => {
  const { data, accessToken: nextToken } = await apiRequestWithAuth<SendCommandResponse>(
    `/api/v1/dev/${deviceId}/cmd`,
    accessToken,
    {
      method: 'POST',
      body: JSON.stringify({
        type,
        params,
        timeoutMs,
      }),
    },
    refreshToken,
    onUnauthorized
  );
  return { ack: data.ack, accessToken: nextToken };
};
