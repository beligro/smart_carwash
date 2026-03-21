import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import styled from 'styled-components';
import WebApiService from '../../shared/services/WebApiService';

const Page = styled.div`
  min-height: 100vh;
  width: 100vw;
  margin: 0;
  padding: 24px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #e0eaf0 0%, #e8eef3 50%, #f0f4f8 100%);
  background-attachment: fixed;
`;

const Card = styled.div`
  background: white;
  border-radius: 12px;
  padding: 28px;
  width: 100%;
  max-width: 400px;
  box-sizing: border-box;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
`;

const Title = styled.h1`
  margin: 0 0 24px;
  font-size: 1.5rem;
  color: #333;
  text-align: center;
`;

const Input = styled.input`
  width: 100%;
  box-sizing: border-box;
  padding: 12px 14px;
  font-size: 16px;
  border: 1px solid #ddd;
  border-radius: 8px;
  margin-bottom: 12px;
  &:focus {
    outline: none;
    border-color: #2481cc;
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
  font-size: 18px;
  line-height: 1;
  opacity: 0.65;
  &:hover {
    opacity: 1;
  }
`;

const Button = styled.button`
  width: 100%;
  padding: 14px;
  font-size: 16px;
  font-weight: 600;
  color: white;
  background: #2481cc;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  margin-top: 8px;
  &:hover:not(:disabled) {
    background: #1a6ba8;
  }
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const Error = styled.p`
  color: #c62828;
  font-size: 14px;
  margin: 8px 0 0;
`;

const LinkBlock = styled(Link)`
  display: block;
  margin-top: 16px;
  padding: 0;
  border: none;
  background: none;
  color: #2481cc;
  font-size: 14px;
  cursor: pointer;
  text-decoration: underline;
  text-align: center;
  width: 100%;
`;

const MIN_PASSWORD_LENGTH = 8;

const WebForgotPasswordPage = () => {
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [newPasswordConfirm, setNewPasswordConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [cooldown, setCooldown] = useState(0);
  const [showPw, setShowPw] = useState(false);
  const [showPwConfirm, setShowPwConfirm] = useState(false);

  useEffect(() => {
    if (step !== 2 || cooldown <= 0) return;
    const t = setInterval(() => setCooldown((c) => (c <= 1 ? 0 : c - 1)), 1000);
    return () => clearInterval(t);
  }, [step, cooldown]);

  const handleSendCode = async (e) => {
    e.preventDefault();
    setError('');
    const em = email.trim().toLowerCase();
    if (!em) {
      setError('Введите email');
      return;
    }
    setLoading(true);
    try {
      await WebApiService.changePasswordSendCode(em);
      setStep(2);
      setCooldown(60);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Не удалось отправить код');
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = async () => {
    const em = email.trim().toLowerCase();
    if (!em || cooldown > 0) return;
    setError('');
    setLoading(true);
    try {
      await WebApiService.changePasswordSendCode(em);
      setCooldown(60);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Ошибка');
    } finally {
      setLoading(false);
    }
  };

  const handleConfirm = async (e) => {
    e.preventDefault();
    setError('');
    const em = email.trim().toLowerCase();
    if (newPassword.length < MIN_PASSWORD_LENGTH) {
      setError(`Пароль не менее ${MIN_PASSWORD_LENGTH} символов`);
      return;
    }
    if (newPassword !== newPasswordConfirm) {
      setError('Пароли не совпадают');
      return;
    }
    if (!code.trim()) {
      setError('Введите код из письма');
      return;
    }
    setLoading(true);
    try {
      await WebApiService.changePasswordConfirm(em, code.trim(), newPassword, newPasswordConfirm);
      alert('Пароль успешно изменён. Введите новый пароль для входа.');
      navigate('/web/login');
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Ошибка смены пароля');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Page>
      <Card>
        <Title>H2O</Title>
        <p style={{ fontSize: 16, fontWeight: 600, marginBottom: 16 }}>Восстановление пароля</p>

        {step === 1 && (
          <form onSubmit={handleSendCode}>
            <Input
              type="email"
              placeholder="Email для восстановления"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              autoFocus
            />
            {error && <Error>{error}</Error>}
            <Button type="submit" disabled={loading}>
              {loading ? 'Отправка…' : 'Отправить код на email'}
            </Button>
          </form>
        )}

        {step === 2 && (
          <form onSubmit={handleConfirm}>
            <Input
              type="text"
              placeholder="Код из письма"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              maxLength={6}
              autoFocus
            />
            <PasswordWrap>
              <InputWithEye
                type={showPw ? 'text' : 'password'}
                placeholder="Новый пароль"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                autoComplete="new-password"
              />
              <EyeBtn type="button" onClick={() => setShowPw((s) => !s)} aria-label={showPw ? 'Скрыть пароль' : 'Показать пароль'}>
                {showPw ? '🙈' : '👁'}
              </EyeBtn>
            </PasswordWrap>
            <PasswordWrap>
              <InputWithEye
                type={showPwConfirm ? 'text' : 'password'}
                placeholder="Подтверждение нового пароля"
                value={newPasswordConfirm}
                onChange={(e) => setNewPasswordConfirm(e.target.value)}
                autoComplete="new-password"
              />
              <EyeBtn type="button" onClick={() => setShowPwConfirm((s) => !s)} aria-label={showPwConfirm ? 'Скрыть пароль' : 'Показать пароль'}>
                {showPwConfirm ? '🙈' : '👁'}
              </EyeBtn>
            </PasswordWrap>
            {error && <Error>{error}</Error>}
            <Button type="submit" disabled={loading}>
              {loading ? 'Сохранение…' : 'Изменить пароль'}
            </Button>
            <Button type="button" disabled={loading || cooldown > 0} onClick={handleResendCode} style={{ background: '#666', marginTop: 8 }}>
              {cooldown > 0 ? `Повторная отправка через ${cooldown} с` : 'Отправить код ещё раз'}
            </Button>
          </form>
        )}

        <LinkBlock to="/web/login">Вернуться ко входу</LinkBlock>
        <LinkBlock to="/web/login" state={{ tab: 'register' }}>Нет аккаунта? Зарегистрироваться</LinkBlock>
      </Card>
    </Page>
  );
};

export default WebForgotPasswordPage;
