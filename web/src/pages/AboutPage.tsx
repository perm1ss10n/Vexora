import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export function AboutPage() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-16">
      <div className="mb-10 max-w-2xl space-y-4">
        <h2 className="text-3xl font-semibold">О платформе KONYX</h2>
        <p className="text-muted-foreground">
          KONYX создаёт премиальные инструменты для операторов connected hardware. Контроль, безопасность и
          оперативность — на первом месте.
        </p>
      </div>
      <div className="grid gap-6 md:grid-cols-3">
        {['Security first', 'Precision telemetry', 'Scalable operations'].map((title) => (
          <Card key={title}>
            <CardHeader>
              <CardTitle>{title}</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              Мы проектируем интерфейсы, которые выдерживают нагрузку и дают мгновенную ясность операторам.
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
