import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import WebApiService from '../../shared/services/WebApiService';

const Page = styled.div`
  padding: 24px;
  max-width: 480px;
  margin: 0 auto;
  @media (min-width: 600px) {
    padding: 32px 40px;
    max-width: 520px;
  }
`;

const Title = styled.h2`
  margin: 0 0 20px;
  font-size: 1.25rem;
  @media (min-width: 600px) {
    font-size: 1.35rem;
  }
`;

const Section = styled.section`
  margin-bottom: 28px;
`;

const Button = styled.button`
  padding: 12px 20px;
  font-size: 16px;
  background: #2481cc;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  margin-top: 12px;
  transition: background 0.2s;
  &:hover:not(:disabled) {
    background: #1a6ba8;
  }
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
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

const LinkBlock = styled.div`
  margin-top: 16px;
  padding: 12px;
  background: #f0f0f0;
  border-radius: 8px;
  word-break: break-all;
  font-size: 14px;
`;

const CopyButton = styled.button`
  margin-top: 8px;
  padding: 8px 16px;
  font-size: 14px;
  background: #333;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
`;

const Error = styled.p`
  color: #c62828;
  font-size: 14px;
  margin: 8px 0 0;
`;

const MIN_PASSWORD_LENGTH = 8;

const WebSettingsPage = ({ user, onLogout }) => {
  const navigate = useNavigate();
  const [link, setLink] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Смена пароля
  const [pwStep, setPwStep] = useState(1);
  const [pwEmail, setPwEmail] = useState('');
  const [pwCode, setPwCode] = useState('');
  const [pwNew, setPwNew] = useState('');
  const [pwConfirm, setPwConfirm] = useState('');
  const [pwLoading, setPwLoading] = useState(false);
  const [pwError, setPwError] = useState('');
  const [pwResendCooldown, setPwResendCooldown] = useState(0);
  const [showPwNew, setShowPwNew] = useState(false);
  const [showPwConfirm, setShowPwConfirm] = useState(false);

  useEffect(() => {
    if (user?.email) setPwEmail(user.email);
  }, [user?.email]);

  useEffect(() => {
    if (pwStep !== 2 || pwResendCooldown <= 0) return;
    const t = setInterval(() => setPwResendCooldown((p) => (p <= 1 ? 0 : p - 1)), 1000);
    return () => clearInterval(t);
  }, [pwStep, pwResendCooldown]);

  const handleRequestLink = async () => {
    setError('');
    setLoading(true);
    try {
      const res = await WebApiService.linkTelegramRequest();
      setLink(res?.link || '');
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Ошибка');
    } finally {
      setLoading(false);
    }
  };

  const copyLink = () => {
    if (link) {
      navigator.clipboard.writeText(link);
      alert('Ссылка скопирована');
    }
  };

  const handleLogout = async () => {
    await WebApiService.logout();
    if (onLogout) onLogout();
    else {
      navigate('/web/login', { replace: true });
      window.location.reload();
    }
  };

  const handlePasswordSendCode = async (e) => {
    e.preventDefault();
    setPwError('');
    const em = (pwEmail || user?.email || '').trim().toLowerCase();
    if (!em) {
      setPwError('Введите email');
      return;
    }
    setPwLoading(true);
    try {
      await WebApiService.changePasswordSendCode(em);
      setPwStep(2);
      setPwResendCooldown(60);
    } catch (err) {
      setPwError(err.response?.data?.error || err.message || 'Не удалось отправить код');
    } finally {
      setPwLoading(false);
    }
  };

  const handlePasswordConfirm = async (e) => {
    e.preventDefault();
    setPwError('');
    const em = (pwEmail || user?.email || '').trim().toLowerCase();
    if (pwNew.length < MIN_PASSWORD_LENGTH) {
      setPwError(`Пароль не менее ${MIN_PASSWORD_LENGTH} символов`);
      return;
    }
    if (pwNew !== pwConfirm) {
      setPwError('Пароли не совпадают');
      return;
    }
    setPwLoading(true);
    try {
      await WebApiService.changePasswordConfirm(em, pwCode.trim(), pwNew, pwConfirm);
      setPwError('');
      setPwStep(1);
      setPwCode('');
      setPwNew('');
      setPwConfirm('');
      alert('Пароль успешно изменён');
    } catch (err) {
      setPwError(err.response?.data?.error || err.message || 'Ошибка смены пароля');
    } finally {
      setPwLoading(false);
    }
  };

  return (
    <Page>
      <Title>Настройки</Title>

      <Section>
        <p>Привяжите аккаунт к Telegram, чтобы получать уведомления и открывать мойку из бота.</p>
        <Button onClick={handleRequestLink} disabled={loading}>
          {loading ? 'Загрузка…' : 'Привязать Telegram'}
        </Button>
        {error && <Error>{error}</Error>}
        {link && (
          <>
            <LinkBlock>
              Откройте ссылку в Telegram (на телефоне или в приложении):
              <br />
              <a href={link} target="_blank" rel="noopener noreferrer">{link}</a>
            </LinkBlock>
            <CopyButton onClick={copyLink}>Скопировать ссылку</CopyButton>
          </>
        )}
      </Section>

      <Section>
        <h3 style={{ margin: '0 0 12px', fontSize: '1rem' }}>Смена пароля</h3>
        {pwStep === 1 ? (
          <form onSubmit={handlePasswordSendCode}>
            <Input
              type="email"
              placeholder="Email"
              value={pwEmail}
              onChange={(e) => setPwEmail(e.target.value)}
              autoComplete="email"
            />
            {pwError && <Error>{pwError}</Error>}
            <Button type="submit" disabled={pwLoading}>
              {pwLoading ? 'Отправка…' : 'Отправить код на email'}
            </Button>
          </form>
        ) : (
          <form onSubmit={handlePasswordConfirm}>
            <Input
              type="text"
              placeholder="Код из письма"
              value={pwCode}
              onChange={(e) => setPwCode(e.target.value)}
              maxLength={6}
            />
            <PasswordWrap>
              <InputWithEye
                type={showPwNew ? 'text' : 'password'}
                placeholder="Новый пароль (не менее 8 символов)"
                value={pwNew}
                onChange={(e) => setPwNew(e.target.value)}
                autoComplete="new-password"
              />
              <EyeBtn
                type="button"
                onClick={() => setShowPwNew((s) => !s)}
                aria-label={showPwNew ? 'Скрыть пароль' : 'Показать пароль'}
              >
                {showPwNew ? '🙈' : '👁'}
              </EyeBtn>
            </PasswordWrap>
            <PasswordWrap>
              <InputWithEye
                type={showPwConfirm ? 'text' : 'password'}
                placeholder="Подтверждение пароля"
                value={pwConfirm}
                onChange={(e) => setPwConfirm(e.target.value)}
                autoComplete="new-password"
              />
              <EyeBtn
                type="button"
                onClick={() => setShowPwConfirm((s) => !s)}
                aria-label={showPwConfirm ? 'Скрыть пароль' : 'Показать пароль'}
              >
                {showPwConfirm ? '🙈' : '👁'}
              </EyeBtn>
            </PasswordWrap>
            {pwError && <Error>{pwError}</Error>}
            <Button type="submit" disabled={pwLoading}>
              {pwLoading ? 'Сохранение…' : 'Изменить пароль'}
            </Button>
            <Button
              type="button"
              onClick={async () => {
                const em = (pwEmail || user?.email || '').trim().toLowerCase();
                if (!em || pwResendCooldown > 0) return;
                setPwLoading(true);
                setPwError('');
                try {
                  await WebApiService.changePasswordSendCode(em);
                  setPwResendCooldown(60);
                } catch (err) {
                  setPwError(err.response?.data?.error || err.message || 'Ошибка');
                } finally {
                  setPwLoading(false);
                }
              }}
              disabled={pwLoading || pwResendCooldown > 0}
              style={{ background: '#666', marginLeft: 8 }}
            >
              {pwResendCooldown > 0 ? `Повтор через ${pwResendCooldown} с` : 'Отправить код повторно'}
            </Button>
            <Button type="button" onClick={() => { setPwStep(1); setPwCode(''); setPwError(''); }} disabled={pwLoading} style={{ background: '#555', marginTop: 8 }}>
              Изменить email
            </Button>
          </form>
        )}
      </Section>

      <Section>
        <Button onClick={handleLogout} style={{ background: '#c62828' }}>
          Выйти из аккаунта
        </Button>
      </Section>

      <Button onClick={() => navigate('/web')} style={{ background: '#666', marginTop: 24 }}>
        Назад
      </Button>
    </Page>
  );
};

export default WebSettingsPage;
