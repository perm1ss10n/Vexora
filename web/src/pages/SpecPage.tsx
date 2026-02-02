import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const specs = [
  {
    title: 'Device orchestration',
    details: ['Статус онлайн/оффлайн', 'История телеметрии', 'Команды управления'],
  },
  {
    title: 'Telemetry intelligence',
    details: ['60 точек за час', 'Метрики в real-time', 'Фильтры и графики'],
  },
  {
    title: 'Secure auth',
    details: ['Mock auth', 'Local storage', 'Role ready layout'],
  },
];

export function SpecPage() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-16">
      <div className="mb-10 max-w-2xl space-y-4">
        <h2 className="text-3xl font-semibold">Спецификация Control Panel</h2>
        <p className="text-muted-foreground">
          Готовый UI-слой под интеграцию с API — для команд, телеметрии и мониторинга.
        </p>
      </div>
      <div className="grid gap-6 md:grid-cols-3">
        {specs.map((spec) => (
          <Card key={spec.title}>
            <CardHeader>
              <CardTitle>{spec.title}</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2 text-sm text-muted-foreground">
                {spec.details.map((detail) => (
                  <li key={detail}>• {detail}</li>
                ))}
              </ul>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
