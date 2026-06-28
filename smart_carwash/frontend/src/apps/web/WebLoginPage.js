import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import styled from 'styled-components';
import WebApiService from '../../shared/services/WebApiService';
import logoUrl from './assets/h2o-logo.webp';

const Page = styled.div`
  min-height: 100vh;
  width: 100vw;
  margin: 0;
  padding: 24px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  background:
    radial-gradient(1200px 600px at 80% -10%, rgba(25,217,255,0.20), transparent 60%),
    radial-gradient(900px 500px at 0% 10%, rgba(80,120,255,0.14), transparent 60%),
    linear-gradient(180deg, #0b1a2b 0%, #0a1626 100%);
  background-attachment: fixed;
  color: #f3f7ff;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
`;

const Header = styled.header`
  width: 100%;
  max-width: 480px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 28px;
`;

const LogoLink = styled.a`
  display: inline-flex;
  align-items: center;
  text-decoration: none;
`;

const LogoImg = styled.img`
  height: 48px;
  width: auto;
  display: block;
`;

const HomeLink = styled.a`
  color: #9fb2c8;
  text-decoration: none;
  font-size: 0.95rem;
  font-weight: 600;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 999px;
  padding: 8px 16px;
  background: rgba(255,255,255,0.04);
  transition: color 0.2s, background 0.2s;
  &:hover {
    color: #f3f7ff;
    background: rgba(255,255,255,0.08);
  }
`;

const Card = styled.div`
  background: #13243a;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 16px;
  padding: 28px;
  width: 100%;
  max-width: 400px;
  box-sizing: border-box;
  box-shadow: 0 10px 30px -10px rgba(0,0,0,0.5);
`;

const Title = styled.h1`
  margin: 0 0 24px;
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: #f3f7ff;
  text-align: center;
`;

const Tabs = styled.div`
  display: flex;
  margin-bottom: 20px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
`;

const Tab = styled.button`
  flex: 1;
  padding: 10px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 15px;
  color: ${p => p.$active ? '#19d9ff' : '#9fb2c8'};
  font-weight: ${p => p.$active ? 600 : 400};
  border-bottom: 2px solid ${p => p.$active ? '#19d9ff' : 'transparent'};
  margin-bottom: -1px;
`;

const Input = styled.input`
  width: 100%;
  box-sizing: border-box;
  padding: 12px 14px;
  font-size: 16px;
  color: #f3f7ff;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 10px;
  margin-bottom: 12px;
  &::placeholder {
    color: #7e91a8;
  }
  &:focus {
    outline: none;
    border-color: #19d9ff;
    box-shadow: 0 0 0 3px rgba(25,217,255,0.18);
  }
`;

const PasswordWrap = styled.div`
  position: relative;
  margin-bottom: 12px;
`;

const InputWithEye = styled(Input)`
  padding-right: 44px;
  margin-bottom: 0;
`;

const EyeBtn = styled.button`
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 6px;
  line-height: 0;
  color: #9fb2c8;
  opacity: 0.85;
  &:hover {
    opacity: 1;
    color: #f3f7ff;
  }
`;

const EyeIcon = ({ off }) => (
  off ? (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
      <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
      <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
      <line x1="2" y1="2" x2="22" y2="22" />
    </svg>
  ) : (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  )
);

const Button = styled.button`
  width: 100%;
  padding: 14px;
  font-size: 16px;
  font-weight: 700;
  color: #0b1a2b;
  background: linear-gradient(135deg, #19d9ff 0%, #0aa5e8 100%);
  border: none;
  border-radius: 10px;
  cursor: pointer;
  margin-top: 8px;
  transition: transform 0.2s, box-shadow 0.2s, opacity 0.2s;
  &:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 20px 60px -20px rgba(25,217,255,0.7);
  }
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const SecondaryButton = styled(Button)`
  color: #f3f7ff;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.12);
  &:hover:not(:disabled) {
    box-shadow: none;
    background: rgba(255,255,255,0.1);
  }
`;

const Error = styled.p`
  color: #ff8a8a;
  font-size: 14px;
  margin: 8px 0 0;
`;

const GuestButton = styled.button`
  width: 100%;
  max-width: 400px;
  padding: 20px;
  font-size: 1.2rem;
  font-weight: 800;
  color: #0b1a2b;
  background: linear-gradient(135deg, #19d9ff 0%, #0aa5e8 100%);
  border: none;
  border-radius: 16px;
  cursor: pointer;
  box-shadow: 0 20px 60px -20px rgba(25,217,255,0.7);
  transition: transform 0.2s, box-shadow 0.2s;
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 24px 70px -18px rgba(25,217,255,0.85);
  }
`;

const GuestHint = styled.p`
  max-width: 400px;
  text-align: center;
  color: #9fb2c8;
  font-size: 14px;
  margin: 10px 0 0;
`;

const Divider = styled.div`
  max-width: 400px;
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  color: #7e91a8;
  font-size: 13px;
  margin: 24px 0;
  &::before, &::after {
    content: '';
    flex: 1;
    height: 1px;
    background: rgba(255,255,255,0.1);
  }
`;

const StatusLink = styled.a`
  display: block;
  width: 100%;
  max-width: 400px;
  box-sizing: border-box;
  margin-top: 14px;
  padding: 14px 20px;
  text-align: center;
  text-decoration: none;
  font-size: 1rem;
  font-weight: 600;
  color: #19d9ff;
  background: rgba(25,217,255,0.08);
  border: 1px solid rgba(25,217,255,0.3);
  border-radius: 12px;
  transition: background 0.2s, color 0.2s;
  &:hover {
    background: rgba(25,217,255,0.14);
    color: #6fe6ff;
    text-decoration: none;
  }
`;

const RegHint = styled.p`
  max-width: 400px;
  text-align: center;
  color: #9fb2c8;
  font-size: 13px;
  line-height: 1.5;
  margin: 16px 0 0;
`;

const MIN_PASSWORD_LENGTH = 8;

const WebLoginPage = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [mode, setMode] = useState(location.state?.tab === 'register' ? 'register' : 'login');
  const [registerStep, setRegisterStep] = useState(1);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [resendCooldown, setResendCooldown] = useState(0);
  const [showPassword, setShowPassword] = useState(false);
  const [showPasswordConfirm, setShowPasswordConfirm] = useState(false);
  const [lastRegisterEmail, setLastRegisterEmail] = useState('');
  const [lastRegisterSentAt, setLastRegisterSentAt] = useState(0);
  const REGISTER_CODE_TTL_SEC = 15 * 60;

  useEffect(() => {
    if (registerStep !== 2 || resendCooldown <= 0) return;
    const t = setInterval(() => {
      setResendCooldown((prev) => (prev <= 1 ? 0 : prev - 1));
    }, 1000);
    return () => clearInterval(t);
  }, [registerStep, resendCooldown]);

  const handleLogin = async (e) => {
    e.preventDefault();
    setError('');
    const em = email.trim().toLowerCase();
    if (!em) {
      setError('Введите email');
      return;
    }
    if (!password) {
      setError('Введите пароль');
      return;
    }
    setLoading(true);
    try {
      const res = await WebApiService.login(em, password);
      localStorage.setItem('web_token', res.token);
      localStorage.setItem('web_user', JSON.stringify(res.user));
      navigate('/web', { replace: true });
      window.location.reload();
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  const handleRegisterSendCode = async (e) => {
    e.preventDefault();
    setError('');
    const em = email.trim().toLowerCase();
    if (!em) {
      setError('Введите email');
      return;
    }
    if (password.length < MIN_PASSWORD_LENGTH) {
      setError(`Пароль не менее ${MIN_PASSWORD_LENGTH} символов`);
      return;
    }
    if (password !== passwordConfirm) {
      setError('Пароли не совпадают');
      return;
    }
    const now = Date.now();
    const elapsed = Math.floor((now - lastRegisterSentAt) / 1000);
    if (em === lastRegisterEmail && lastRegisterSentAt > 0 && elapsed < REGISTER_CODE_TTL_SEC) {
      setRegisterStep(2);
      setResendCooldown(elapsed >= 60 ? 0 : 60 - elapsed);
      return;
    }
    setLoading(true);
    try {
      await WebApiService.registerSendCode(em, password, passwordConfirm);
      setLastRegisterEmail(em);
      setLastRegisterSentAt(now);
      setRegisterStep(2);
      setResendCooldown(60);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Не удалось отправить код');
    } finally {
      setLoading(false);
    }
  };

  const handleRegisterVerify = async (e) => {
    e.preventDefault();
    setError('');
    if (!code.trim()) {
      setError('Введите код из письма');
      return;
    }
    const em = email.trim().toLowerCase();
    setLoading(true);
    try {
      const res = await WebApiService.registerVerify(em, code.trim(), password, passwordConfirm);
      localStorage.setItem('web_token', res.token);
      localStorage.setItem('web_user', JSON.stringify(res.user));
      navigate('/web', { replace: true });
      window.location.reload();
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Неверный код');
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = async () => {
    if (resendCooldown > 0) return;
    const em = email.trim().toLowerCase();
    setError('');
    setLoading(true);
    try {
      await WebApiService.registerSendCode(em, password, passwordConfirm);
      setResendCooldown(60);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Не удалось отправить код');
    } finally {
      setLoading(false);
    }
  };

  const switchToLogin = () => {
    setMode('login');
    setError('');
    setRegisterStep(1);
    setCode('');
  };

  const switchToRegister = () => {
    setMode('register');
    setError('');
    setRegisterStep(1);
    setCode('');
    setResendCooldown(0);
  };

  return (
    <Page>
      <Header>
        <LogoLink href="/" aria-label="На главную H2O">
          <LogoImg src={logoUrl} alt="H2O — автомойка самообслуживания" />
        </LogoLink>
        <HomeLink href="/">← На главную</HomeLink>
      </Header>

      {/* Приоритетный сценарий — помыть как гость, без регистрации */}
      <GuestButton type="button" onClick={() => navigate('/web/guest')}>
        Помыть машину как гость
      </GuestButton>
      <GuestHint>Без регистрации — выбрали услугу, оплатили и поехали</GuestHint>

      <StatusLink href="/status" target="_blank" rel="noopener noreferrer">
        Статус загруженности в реальном времени
      </StatusLink>

      <Divider>есть учётная запись?</Divider>

      <Card>
        <Title>Вход в аккаунт</Title>
        <Tabs>
          <Tab $active={mode === 'login'} onClick={switchToLogin} type="button">Вход</Tab>
          <Tab $active={mode === 'register'} onClick={switchToRegister} type="button">Регистрация</Tab>
        </Tabs>

        {mode === 'login' && (
          <form onSubmit={handleLogin}>
            <Input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              autoFocus
            />
            <PasswordWrap>
              <InputWithEye
                type={showPassword ? 'text' : 'password'}
                placeholder="Пароль"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
              />
              <EyeBtn type="button" onClick={() => setShowPassword((s) => !s)} aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}>
                <EyeIcon off={showPassword} />
              </EyeBtn>
            </PasswordWrap>
            <Link to="/web/login/forgot" style={{ marginTop: 4, marginBottom: 8, display: 'block', color: '#19d9ff', fontSize: 13, textDecoration: 'underline' }}>
              Забыли пароль?
            </Link>
            {error && <Error>{error}</Error>}
            <Button type="submit" disabled={loading}>
              {loading ? 'Вход…' : 'Войти'}
            </Button>
          </form>
        )}

        {mode === 'register' && registerStep === 1 && (
          <form onSubmit={handleRegisterSendCode}>
            <Input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              autoFocus
            />
            <PasswordWrap>
              <InputWithEye
                type={showPassword ? 'text' : 'password'}
                placeholder="Пароль (не менее 8 символов)"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="new-password"
              />
              <EyeBtn
                type="button"
                onClick={() => setShowPassword((s) => !s)}
                aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
              >
                <EyeIcon off={showPassword} />
              </EyeBtn>
            </PasswordWrap>
            <PasswordWrap>
              <InputWithEye
                type={showPasswordConfirm ? 'text' : 'password'}
                placeholder="Подтверждение пароля"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                autoComplete="new-password"
              />
              <EyeBtn
                type="button"
                onClick={() => setShowPasswordConfirm((s) => !s)}
                aria-label={showPasswordConfirm ? 'Скрыть пароль' : 'Показать пароль'}
              >
                <EyeIcon off={showPasswordConfirm} />
              </EyeBtn>
            </PasswordWrap>
            {error && <Error>{error}</Error>}
            <Button type="submit" disabled={loading}>
              {loading ? 'Отправка…' : 'Отправить код на email'}
            </Button>
          </form>
        )}

        {mode === 'register' && registerStep === 2 && (
          <form onSubmit={handleRegisterVerify}>
            <p style={{ margin: '0 0 12px', fontSize: '14px', color: '#9fb2c8', lineHeight: '1.5' }}>
              Код подтверждения отправлен на <strong style={{ color: '#f3f7ff' }}>{email}</strong>.<br/>
              Если письмо не пришло — проверьте папку <strong style={{ color: '#f3f7ff' }}>Спам</strong>.
            </p>
            <Input
              type="text"
              placeholder="Код из письма"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              autoFocus
              maxLength={6}
            />
            {error && <Error>{error}</Error>}
            <Button type="submit" disabled={loading}>
              {loading ? 'Проверка…' : 'Зарегистрироваться'}
            </Button>
            <SecondaryButton
              type="button"
              onClick={handleResendCode}
              disabled={loading || resendCooldown > 0}
            >
              {resendCooldown > 0 ? `Отправить повторно через ${resendCooldown} сек` : 'Отправить повторно'}
            </SecondaryButton>
            <SecondaryButton type="button" onClick={() => { setRegisterStep(1); setCode(''); setError(''); }} disabled={loading}>
              Изменить email
            </SecondaryButton>
          </form>
        )}
      </Card>

      <RegHint>
        Регистрация не обязательна — помыть можно и без неё.
        Но с аккаунтом удобнее: история моек, уведомления о статусе
        и программа лояльности.
      </RegHint>
    </Page>
  );
};

export default WebLoginPage;
