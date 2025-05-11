import React, { useEffect, useState } from 'react';
import { 
  useMainButton, 
  useBackButton, 
  useTheme, 
  useUser,
  usePopup
} from '@telegram-apps/sdk-react';
import { 
  Container, 
  Card, 
  Title, 
  Subtitle, 
  Text, 
  Grid, 
  BoxItem, 
  LoadingSpinner,
  FlexBetween,
  StatusIndicator
} from '../components/styled';
import { getCarwashInfo, CarwashInfo } from '../services/api';

const HomePage: React.FC = () => {
  const [carwashInfo, setCarwashInfo] = useState<CarwashInfo | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  
  // Use Telegram SDK hooks
  const mainButton = useMainButton();
  const backButton = useBackButton();
  const theme = useTheme();
  const user = useUser();
  const popup = usePopup();
  
  // Initialize Telegram Mini App
  useEffect(() => {
    // Configure main button
    if (mainButton) {
      mainButton.setText('Обновить');
      mainButton.onClick(fetchCarwashInfo);
      mainButton.show();
    }
    
    // Fetch carwash info
    fetchCarwashInfo();
    
    // Cleanup
    return () => {
      if (mainButton) {
        mainButton.hide();
        mainButton.offClick(fetchCarwashInfo);
      }
    };
  }, [mainButton]);
  
  // Fetch carwash info from API
  const fetchCarwashInfo = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getCarwashInfo();
      setCarwashInfo(data);
    } catch (err) {
      console.error('Error fetching carwash info:', err);
      setError('Failed to load carwash information. Please try again.');
      
      // Show error popup
      if (popup) {
        popup.showAlert('Не удалось загрузить информацию об автомойке. Пожалуйста, попробуйте снова.');
      }
    } finally {
      setLoading(false);
    }
  };
  
  // Render loading state
  if (loading) {
    return (
      <Container>
        <Card>
          <Title>Smart Carwash</Title>
          <Text>Загрузка информации об автомойке...</Text>
          <LoadingSpinner />
        </Card>
      </Container>
    );
  }
  
  // Render error state
  if (error) {
    return (
      <Container>
        <Card>
          <Title>Smart Carwash</Title>
          <Text style={{ color: '#F44336' }}>{error}</Text>
        </Card>
      </Container>
    );
  }
  
  return (
    <Container>
      <Card>
        <Title>Smart Carwash</Title>
        <Text>
          Добро пожаловать в сервис Smart Carwash! Проверьте доступность наших моечных боксов ниже.
        </Text>
        
        {carwashInfo && (
          <>
            <FlexBetween style={{ marginBottom: '16px' }}>
              <div>
                <Subtitle>Доступные боксы</Subtitle>
              </div>
              <div>
                <Text style={{ margin: 0 }}>
                  <StatusIndicator status="available" />
                  {carwashInfo.available_boxes} из {carwashInfo.total_boxes}
                </Text>
              </div>
            </FlexBetween>
            
            {carwashInfo.available_boxes > 0 ? (
              <>
                <Text>Следующие боксы сейчас доступны:</Text>
                <Grid>
                  {carwashInfo.available_box_numbers.map(boxNumber => (
                    <BoxItem key={boxNumber} status="available">
                      Бокс {boxNumber}
                    </BoxItem>
                  ))}
                </Grid>
              </>
            ) : (
              <Text>
                Все моечные боксы сейчас заняты. Пожалуйста, проверьте позже.
              </Text>
            )}
          </>
        )}
      </Card>
      
      <Card>
        <Subtitle>Как это работает</Subtitle>
        <Text>
          1. Проверьте доступность моечных боксов.
        </Text>
        <Text>
          2. Посетите нашу автомойку.
        </Text>
        <Text>
          3. Используйте доступный бокс для мойки вашего автомобиля.
        </Text>
        <Text style={{ marginBottom: 0 }}>
          4. Наслаждайтесь чистым автомобилем!
        </Text>
      </Card>
    </Container>
  );
};

export default HomePage;
