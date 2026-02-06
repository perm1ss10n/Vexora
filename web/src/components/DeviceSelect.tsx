import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectViewport } from '@/components/ui/select';
import { useDevices } from '@/features/devices/hooks';

interface DeviceSelectProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function DeviceSelect({ value, onChange, placeholder = 'Select device', className }: DeviceSelectProps) {
  const { data: devices = [] } = useDevices();

  return (
    <div className={className}>
      <Select value={value} onValueChange={onChange} disabled={devices.length === 0}>
        <SelectTrigger>
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent>
          <SelectViewport>
            {devices.map((device) => (
              <SelectItem key={device.deviceId} value={device.deviceId}>
                {device.deviceId}
              </SelectItem>
            ))}
          </SelectViewport>
        </SelectContent>
      </Select>
    </div>
  );
}
