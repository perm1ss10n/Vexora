import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

const steps = [
  'Оставьте заявку на демо-доступ',
  'Уточните архитектуру устройств',
  'Получите персональный onboarding',
];

export function OrderPage() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-16">
      <div className="mb-10 max-w-2xl space-y-4">
        <h2 className="text-3xl font-semibold">Как заказать</h2>
        <p className="text-muted-foreground">
          KONYX сопровождает внедрение от пилота до промышленного масштаба.
        </p>
      </div>
      <div className="grid gap-6 md:grid-cols-3">
        {steps.map((step, index) => (
          <Card key={step}>
            <CardHeader>
              <CardTitle>Шаг {index + 1}</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">{step}</CardContent>
          </Card>
        ))}
      </div>
      <div className="mt-10">
        <Button size="lg">Запросить коммерческое предложение</Button>
      </div>
    </section>
  );
}
