import React, { useState, useEffect } from 'react';
import './LoyaltyProgress.css';
import ApiService from '../../services/ApiService';

const LoyaltyProgress = ({ userId }) => {
  const [progress, setProgress] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!userId) return;
    
    const fetchProgress = async () => {
      try {
        const data = await ApiService.get(`/api/loyalty/progress?user_id=${userId}`);
        setProgress(data);
      } catch (error) {
        console.error('Ошибка загрузки прогресса лояльности:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchProgress();
  }, [userId]);

  if (loading || !progress) return null;

  const currentCount = progress.count;
  const nextFreeAt = progress.next_free_at;
  const isFreeAvailable = progress.is_free_available;
  
  // Прогресс до следующей бесплатной
  const progressToNext = currentCount % 10;
  const progressPercent = (progressToNext / 10) * 100;

  return (
    <div className="loyalty-progress">
      <div className="loyalty-header">
        <span className="loyalty-icon">🎁</span>
        <h3>Программа лояльности</h3>
      </div>

      <div className="loyalty-content">
        {isFreeAvailable ? (
          <div className="free-wash-available">
            <div className="free-wash-badge">
              🎉 БЕСПЛАТНАЯ МОЙКА ДОСТУПНА! 🎉
            </div>
            <p className="free-wash-text">
              Вы помыли {currentCount} раз! Следующая мойка с химией - БЕСПЛАТНО!
            </p>
          </div>
        ) : (
          <>
            <div className="progress-info">
              <span className="progress-count">{progressToNext}/10</span>
              <span className="progress-text">до бесплатной мойки</span>
            </div>

            <div className="progress-bar">
              <div 
                className="progress-bar-fill" 
                style={{ width: `${progressPercent}%` }}
              />
            </div>

            <p className="total-washes">
              Всего помыто: {currentCount} | Следующая бесплатная: №{nextFreeAt}
            </p>
          </>
        )}
      </div>

      <div className="loyalty-hint">
        <small>
          📢 Каждая 10-я мойка с химией - бесплатно (30 мин + 5 мин химии)
        </small>
      </div>
    </div>
  );
};

export default LoyaltyProgress;
