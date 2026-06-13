import React, { useState } from 'react';
import styles from './Header.module.css';

/**
 * Header: шапка приложения.
 * @param {string} theme - 'light' | 'dark'
 * @param {Function} onBack - стрелка «Назад»
 * @param {Function} onLogout - выход
 */
const Header = ({ theme = 'light', onBack, onLogout }) => {
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  const [menuOpen, setMenuOpen] = useState(false);

  const toggleMenu = () => setMenuOpen((prev) => !prev);
  const closeMenu = () => setMenuOpen(false);

  return (
    <header className={`${styles.header} ${themeClass}`}>
      {onBack && (
        <button
          onClick={onBack}
          className={`${styles.backButton} ${themeClass}`}
          type="button"
          aria-label="Назад"
        >
          ←
        </button>
      )}
      <div className={styles.logo}>
        <span className={styles.logoIcon}>🚿</span>
        Автомойка H2O
      </div>
      {onLogout && (
        <>
          <button
            type="button"
            onClick={onLogout}
            className={styles.logoutButtonDesktop}
          >
            Выйти
          </button>
          <button
            type="button"
            onClick={toggleMenu}
            className={styles.menuButtonMobile}
            aria-label="Меню"
          >
            ☰
          </button>
          {menuOpen && (
            <div className={styles.menuOverlay} onClick={closeMenu}>
              <div className={styles.menuContent} onClick={(e) => e.stopPropagation()}>
                <button
                  type="button"
                  className={styles.menuLogoutButton}
                  onClick={() => { closeMenu(); onLogout(); }}
                >
                  Выйти
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </header>
  );
};

export default Header;
