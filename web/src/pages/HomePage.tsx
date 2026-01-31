import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

const highlights = [
  {
    title: 'Unified device fleet view',
    description: 'Сводная картина статусов, прошивок и телеметрии в одном тёмном консоле.',
  },
  {
    title: 'Operational-grade telemetry',
    description: 'Готовые метрики и линия тренда с фильтрами по устройствам.',
  },
  {
    title: 'Instant command dispatch',
    description: 'Отправка команд, контроль ACK и журнал операций.',
  },
];

export function HomePage() {
  return (
    <div className="grid-surface">
      <section className="mx-auto flex max-w-6xl flex-col gap-10 px-6 py-20">
        <div className="max-w-2xl space-y-6">
          <span className="rounded-full border border-border/70 bg-muted/40 px-4 py-2 text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Premium-tech control layer
          </span>
          <h1 className="text-4xl font-semibold leading-tight md:text-5xl">
            KONYX
            <span className="block text-gradient">Operations for connected hardware</span>
          </h1>
          <p className="text-base text-muted-foreground md:text-lg">
            Ускорьте наблюдаемость, управление и диагностику IoT-устройств в безопасной dark-first среде.
          </p>
          <div className="flex flex-wrap gap-4">
            <Button asChild size="lg">
              <Link to="/auth/register">Запросить доступ</Link>
            </Button>
            <Button asChild variant="outline" size="lg">
              <Link to="/spec">Смотреть спецификацию</Link>
            </Button>
          </div>
        </div>
        <div className="grid gap-6 md:grid-cols-3">
          {highlights.map((item) => (
            <Card key={item.title} className="bg-card/80">
              <CardHeader>
                <CardTitle>{item.title}</CardTitle>
                <CardDescription>{item.description}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="h-1 w-12 rounded-full bg-primary/70" />
              </CardContent>
            </Card>
          ))}
        </div>
      </section>
    </div>
  );
}
