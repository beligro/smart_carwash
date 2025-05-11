import React, { useEffect, useState } from 'react';
import { TelegramProvider, useUser } from '@telegram-apps/sdk-react';
import { createUser } from './services/api';
import HomePage from './pages/HomePage';
import { LoadingSpinner, Container, Card, Text } from './components/styled';
import './App.css';

// Wrapper component that uses Telegram SDK hooks
const AppContent: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  
  // Get user data from Telegram SDK
  const user = useUser();
  
  useEffect(() => {
    // Register user on first visit
    const registerUser = async () => {
      try {
        if (user) {
          // Create or get existing user
          await createUser({
            telegram_id: user.id.toString(),
            username: user.username || undefined,
            first_name: user.first_name,
            last_name: user.last_name || undefined
          });
        }
        
        setLoading(false);
      } catch (err) {
        console.error('Error registering user:', err);
        setError('Failed to initialize the application. Please try again.');
        setLoading(false);
      }
    };
    
    registerUser();
  }, [user]);

  // Render loading state
  if (loading) {
    return (
      <Container>
        <Card>
          <Text>Initializing Smart Carwash...</Text>
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
          <Text style={{ color: '#F44336' }}>{error}</Text>
        </Card>
      </Container>
    );
  }

  return <HomePage />;
};

function App() {
  // Get the initData from URL or localStorage
  const initDataRaw = window.Telegram?.WebApp?.initData || '';
  
  return (
    <TelegramProvider initData={initDataRaw}>
      <AppContent />
    </TelegramProvider>
  );
}

export default App;
