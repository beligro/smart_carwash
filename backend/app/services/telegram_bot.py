import logging
from telegram import Update
from telegram.ext import (
    Application,
    CommandHandler,
    ContextTypes,
    MessageHandler,
    filters,
)

from app.utils.config import get_settings
from app.services.user_service import create_user_if_not_exists

# Configure logging
logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO
)
logger = logging.getLogger(__name__)

# Get settings
settings = get_settings()

async def start_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle the /start command."""
    user = update.effective_user
    
    # Create user if not exists
    await create_user_if_not_exists(
        telegram_id=str(user.id),
        username=user.username,
        first_name=user.first_name,
        last_name=user.last_name,
    )
    
    # Create a mini app URL
    mini_app_url = f"https://t.me/{settings.telegram_bot_name}/app"
    
    # Send welcome message with mini app button
    await update.message.reply_text(
        f"Привет, {user.first_name}! Добро пожаловать в Smart Carwash Bot. "
        f"Нажмите на кнопку ниже, чтобы открыть приложение.",
        reply_markup={
            "inline_keyboard": [
                [
                    {
                        "text": "Открыть Smart Carwash",
                        "web_app": {"url": mini_app_url},
                    }
                ]
            ]
        },
    )

async def help_command(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle the /help command."""
    await update.message.reply_text(
        "Это бот для умной автомойки. Используйте команду /start, чтобы начать."
    )

async def handle_message(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle messages."""
    await update.message.reply_text(
        "Используйте команду /start, чтобы начать работу с ботом."
    )

async def start_bot() -> None:
    """Start the Telegram bot."""
    # Create the Application
    application = Application.builder().token(settings.telegram_bot_token).build()

    # Add handlers
    application.add_handler(CommandHandler("start", start_command))
    application.add_handler(CommandHandler("help", help_command))
    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))

    # Start the Bot
    await application.initialize()
    await application.start()
    await application.updater.start_polling()
    
    logger.info("Bot started successfully!")
    
    # Run the bot until the user presses Ctrl-C
    await application.updater.stop()
    await application.stop()
