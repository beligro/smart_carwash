import React, { useState } from 'react';
import styled from 'styled-components';
import { useNavigate, useLocation } from 'react-router-dom';
import { getTheme } from '../../shared/styles/theme';

const FormContainer = styled.div`
  max-width: 400px;
  margin: 0 auto;
  padding: 20px;
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
`;

const Title = styled.h2`
  text-align: center;
  margin-bottom: 20px;
  color: ${props => props.theme.textColor};
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
`;

const FormGroup = styled.div`
  margin-bottom: 15px;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
  color: ${props => props.theme.textColor};
`;

const Input = styled.input`
  width: 100%;
  padding: 10px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  font-size: 16px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  
  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }
`;

const Select = styled.select`
  width: 100%;
  padding: 10px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  font-size: 16px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }
`;

const PasswordWrap = styled.div`
  position: relative;
`;

const EyeButton = styled.button`
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 13px;
  color: ${props => props.theme.primaryColor};
  padding: 4px;
  line-height: 1;
`;

const Button = styled.button`
  padding: 10px 15px;
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: ${props => props.theme.primaryColorHover};
  }
  
  &:disabled {
    background-color: ${props => props.theme.disabledColor};
    cursor: not-allowed;
  }
`;

const ErrorMessage = styled.div`
  color: ${props => props.theme.errorColor};
  margin-top: 15px;
  text-align: center;
`;

/**
 * Компонент формы авторизации
 * @param {Object} props - Свойства компонента
 * @param {string} props.title - Заголовок формы
 * @param {Function} props.onLogin - Функция для авторизации
 * @param {string} props.redirectPath - Путь для перенаправления после успешной авторизации
 * @returns {React.ReactNode} - Форма авторизации
 */
const LoginForm = ({ title, onLogin, redirectPath, usernameOptions }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  
  const navigate = useNavigate();
  const location = useLocation();
  const theme = getTheme('light');
  
  // Получаем путь для перенаправления после успешной авторизации
  const from = location.state?.from?.pathname || redirectPath || '/';
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Проверяем, что поля заполнены
    if (!username || !password) {
      setError('Пожалуйста, заполните все поля');
      return;
    }
    
    setLoading(true);
    setError('');
    
    try {
      // Вызываем функцию авторизации
      const response = await onLogin(username, password);
      
      // Определяем, куда перенаправить пользователя
      let redirectTo = from;
      
      // Если нет сохраненного пути или это главная страница, перенаправляем на соответствующий интерфейс
      if (from === '/' || from === '/admin/login' || from === '/cashier/login') {
        if (response.is_admin) {
          redirectTo = '/admin';
        } else {
          redirectTo = '/cashier';
        }
      }
      
      // Перенаправляем пользователя
      navigate(redirectTo, { replace: true });
    } catch (err) {
      // Обрабатываем ошибку авторизации
      if (err.response && err.response.data && err.response.data.error) {
        setError(err.response.data.error);
      } else {
        setError('Произошла ошибка при авторизации. Пожалуйста, попробуйте снова.');
      }
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <FormContainer theme={theme}>
      <Title theme={theme}>{title}</Title>
      <Form onSubmit={handleSubmit}>
        <FormGroup>
          <Label theme={theme}>Имя пользователя</Label>
          {usernameOptions && usernameOptions.length > 0 ? (
            <Select
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={loading}
              theme={theme}
            >
              <option value="">— выберите себя —</option>
              {usernameOptions.map((u) => (
                <option key={u} value={u}>{u}</option>
              ))}
            </Select>
          ) : (
            <Input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={loading}
              theme={theme}
            />
          )}
        </FormGroup>
        <FormGroup>
          <Label theme={theme}>Пароль</Label>
          <PasswordWrap>
            <Input
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loading}
              theme={theme}
              style={{ paddingRight: 44 }}
            />
            <EyeButton
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              theme={theme}
            >
              {showPassword ? 'Скрыть' : 'Показать'}
            </EyeButton>
          </PasswordWrap>
        </FormGroup>
        <Button type="submit" disabled={loading} theme={theme}>
          {loading ? 'Вход...' : 'Войти'}
        </Button>
        {error && <ErrorMessage theme={theme}>{error}</ErrorMessage>}
      </Form>
    </FormContainer>
  );
};

export default LoginForm;
