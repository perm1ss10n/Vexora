import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { useAuthStore } from '@/store/auth';

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((state) => state.login);
  const status = useAuthStore((state) => state.status);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError('');
    try {
      await login(email, password);
      const redirectTo = (location.state as { from?: Location })?.from?.pathname ?? '/app';
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось войти.');
    }
  };

  return (
    <div className="grid min-h-screen place-items-center bg-background px-6">
      <Card className="w-full max-w-md border border-border/60 bg-card/80 shadow-glow">
        <CardHeader>
          <CardTitle>Войти в KONYX</CardTitle>
          <CardDescription>Доступ к панели управления устройствами.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <Input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
            <Input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
            {error ? <p className="text-sm text-rose-400">{error}</p> : null}
            <Button type="submit" className="w-full" disabled={status === 'loading'}>
              {status === 'loading' ? 'Входим...' : 'Войти'}
            </Button>
            <p className="text-center text-xs text-muted-foreground">
              Нет доступа?{' '}
              <Link className="text-foreground underline-offset-4 hover:underline" to="/register">
                Создать аккаунт
              </Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
