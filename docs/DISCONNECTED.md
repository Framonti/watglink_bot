# Disconnected - What?

If you're reading this page, you're probably trying to figure out why the bot sent you an error saying that your device is disconnected. Here's why this error exists, why it happens, and how to fix it.

Skip to: [How to fix](#how-can-i-fix-this-error?)

### Why does this error exist?

Whatsapp-Telegram Linker relies completely on the web.whatsapp.com API. Our servers act like your computer, connected to WhatsApp Web, with the difference being that instead of showing your chats on a static web page, they're shown on Telegram.

WhatsApp Web works in this way:

![How WhatsApp Web works](https://imgur.com/WZx0K7G.png)

because that's the only way with WhatsApp End-To-End Encryption to get readable data: always using your phone for decryption.

It's easy to understand how your phone is necessary. Even with WhatsApp Web, if your phone is disconnected, you are not be able to send and receive messages or any kind of data.

![](https://imgur.com/ZsVx2fe.png)

### Why does this error happen?

You might be wondering: why does this error happen? If my phone is connected to the internet, shouldn't it be able to send data to WhatsApp Web - Or in this case to Whatsapp-Telegram Linker Servers?

Sadly, it's not as easy as that.

Often, phone manufacturers use weird "optimization" algorithms that kill apps running in the background to save battery. If this is the case, WhatsApp would be able to send and receive data to our servers only when it's open in the foreground or for a short period of time after that.

### How can I fix this error?

First of all, check that your phone is really connected, your device is not in airplane/flight mode and you have a decent signal coverage.

If the error persists, select the phone manufacturer from the list below by clicking on its name, and follow the procedures described in the article to stop your operating system from killing WhatsApp. (You will be redirected to an external site)

[Huawei](https://www.geekdashboard.com/stop-android-killing-apps-background/#huawei) - [Samsung](https://www.geekdashboard.com/stop-android-killing-apps-background/#samsung) - [OnePlus](https://www.geekdashboard.com/stop-android-killing-apps-background/#oneplus) - [Asus](https://www.geekdashboard.com/stop-android-killing-apps-background/#asus) - [Nokia](https://www.geekdashboard.com/stop-android-killing-apps-background/#nokia) - [Lenovo](https://www.geekdashboard.com/stop-android-killing-apps-background/#lenovo) - [Xiaomi](https://www.geekdashboard.com/stop-android-killing-apps-background/#xiaomi) - [Google, Motorola, Others](https://www.geekdashboard.com/stop-android-killing-apps-background/#others)

If you couldn't find a fix, feel free to open a issue.