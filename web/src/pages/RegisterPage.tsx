import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { useAuthStore } from '@/store/auth';

export function RegisterPage() {
  const navigate = useNavigate();
  const register = useAuthStore((state) => state.register);
  const status = useAuthStore((state) => state.status);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (password !== confirm) {
      setError('Пароли не совпадают.');
      return;
    }
    setError('');
    try {
      await register(email, password);
      navigate('/app', { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать аккаунт.');
    }
  };

  return (
    <div className="grid min-h-screen place-items-center bg-background px-6">
      <Card className="w-full max-w-md border border-border/60 bg-card/80 shadow-glow">
        <CardHeader>
          <CardTitle>Регистрация</CardTitle>
          <CardDescription>Создайте доступ в KONYX Control Panel.</CardDescription>
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
            <Input
              type="password"
              placeholder="Confirm password"
              value={confirm}
              onChange={(event) => setConfirm(event.target.value)}
              required
            />
            {error ? <p className="text-sm text-rose-400">{error}</p> : null}
            <Button type="submit" className="w-full" disabled={status === 'loading'}>
              {status === 'loading' ? 'Создаем...' : 'Зарегистрироваться'}
            </Button>
            <p className="text-center text-xs text-muted-foreground">
              Уже есть доступ?{' '}
              <Link className="text-foreground underline-offset-4 hover:underline" to="/login">
                Войти
              </Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
