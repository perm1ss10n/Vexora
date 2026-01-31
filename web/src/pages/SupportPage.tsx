import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const supportCards = [
  { title: '24/7 Monitoring', detail: 'Наблюдение за критичными узлами и SLA 99.9%.' },
  { title: 'Incident response', detail: 'Готовые сценарии реагирования и журнал команд.' },
  { title: 'Dedicated success', detail: 'Персональный инженер и регулярные отчёты.' },
];

export function SupportPage() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-16">
      <div className="mb-10 max-w-2xl space-y-4">
        <h2 className="text-3xl font-semibold">Поддержка</h2>
        <p className="text-muted-foreground">
          Экспертная команда KONYX обеспечивает сопровождение на всех этапах.
        </p>
      </div>
      <div className="grid gap-6 md:grid-cols-3">
        {supportCards.map((item) => (
          <Card key={item.title}>
            <CardHeader>
              <CardTitle>{item.title}</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">{item.detail}</CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
