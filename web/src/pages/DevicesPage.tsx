import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useDevices } from '@/features/devices/hooks';

export function DevicesPage() {
  const navigate = useNavigate();
  const { data = [], isLoading } = useDevices();
  const [query, setQuery] = useState('');

  const filtered = useMemo(() => {
    const lower = query.toLowerCase();
    return data.filter((device) => device.deviceId.toLowerCase().includes(lower));
  }, [data, query]);

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <CardTitle>Devices</CardTitle>
          <Input
            placeholder="Search by deviceId"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            className="max-w-xs"
          />
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Loading devices...</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Device ID</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last seen</TableHead>
                  <TableHead>Last telemetry</TableHead>
                  <TableHead>FW version</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((device) => (
                  <TableRow
                    key={device.deviceId}
                    className="cursor-pointer"
                    onClick={() => navigate(`/app/devices/${device.deviceId}`)}
                  >
                    <TableCell className="font-medium text-foreground">{device.deviceId}</TableCell>
                    <TableCell>
                      <Badge variant={device.status === 'online' ? 'success' : 'danger'}>
                        {device.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(device.lastSeen).toLocaleString('ru-RU')}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(device.lastTelemetryTs).toLocaleString('ru-RU')}
                    </TableCell>
                    <TableCell>{device.fwVersion}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
