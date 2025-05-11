import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { useSDK } from '@tma.js/sdk-react';
import CarwashBox from '../components/CarwashBox';
import LoadingScreen from '../components/LoadingScreen';
import ErrorScreen from '../components/ErrorScreen';
import api from '../services/api';

const HomeContainer = styled.div`
  padding: 16px;
`;

const Header = styled.div`
  text-align: center;
  margin-bottom: 24px;
`;

const Title = styled.h1`
  font-size: 24px;
  margin-bottom: 8px;
`;

const Subtitle = styled.p`
  font-size: 16px;
  color: var(--tg-theme-hint-color, #6c757d);
  margin: 0;
`;

const StatsContainer = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-bottom: 24px;
`;

const StatCard = styled.div`
  background-color: var(--tg-theme-secondary-bg-color, white);
  border-radius: 8px;
  padding: 12px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
`;

const StatValue = styled.div`
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 4px;
`;

const StatLabel = styled.div`
  font-size: 12px;
  color: var(--tg-theme-hint-color, #6c757d);
`;

const BoxesContainer = styled.div`
  margin-top: 20px;
`;

const BoxesTitle = styled.h2`
  font-size: 18px;
  margin-bottom: 16px;
`;

const RefreshButton = styled.button`
  width: 100%;
  margin-top: 20px;
`;

const HomePage = () => {
  const { user } = useSDK();
  const [carwashInfo, setCarwashInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchCarwashInfo = async () => {
    setLoading(true);
    try {
      const data = await api.carwash.getInfo();
      setCarwashInfo(data);
      setError(null);
    } catch (err) {
      console.error('Error fetching carwash info:', err);
      setError('Не удалось загрузить информацию о мойке');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCarwashInfo();
    
    // Register user if Telegram user data is available
    const registerUser = async () => {
      if (user) {
        try {
          await api.users.ensureExists({
            telegram_id: user.id.toString(),
            username: user.username,
            first_name: user.firstName,
            last_name: user.lastName,
          });
        } catch (err) {
          console.error('Error registering user:', err);
        }
      }
    };
    
    registerUser();
  }, [user]);

  const handleRefresh = () => {
    fetchCarwashInfo();
  };

  if (loading) {
    return <LoadingScreen message="Загрузка информации о мойке..." />;
  }

  if (error) {
    return <ErrorScreen message={error} />;
  }

  return (
    <HomeContainer>
      <Header>
        <Title>Smart Carwash</Title>
        <Subtitle>Добро пожаловать в умную автомойку!</Subtitle>
      </Header>

      {carwashInfo && (
        <>
          <StatsContainer>
            <StatCard>
              <StatValue>{carwashInfo.total_boxes}</StatValue>
              <StatLabel>Всего боксов</StatLabel>
            </StatCard>
            <StatCard>
              <StatValue>{carwashInfo.available_boxes}</StatValue>
              <StatLabel>Доступно</StatLabel>
            </StatCard>
            <StatCard>
              <StatValue>{carwashInfo.occupied_boxes}</StatValue>
              <StatLabel>Занято</StatLabel>
            </StatCard>
          </StatsContainer>

          <BoxesContainer>
            <BoxesTitle>Статус боксов</BoxesTitle>
            {carwashInfo.boxes.map((box) => (
              <CarwashBox key={box.id} box={box} />
            ))}
          </BoxesContainer>

          <RefreshButton onClick={handleRefresh}>
            Обновить информацию
          </RefreshButton>
        </>
      )}
    </HomeContainer>
  );
};

export default HomePage;
