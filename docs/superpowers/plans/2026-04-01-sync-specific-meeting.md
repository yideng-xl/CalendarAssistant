# 特定日程快速同步 (2026-04-02 评审会议) 执行计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将特定的评审会议日程同步至 macOS 日历，并设置 15/5/0 分钟三级提醒。

**Architecture:** 编写一个临时的 Python 脚本，复用 `CalendarAssistant/calendar_sync_prototype.py` 中的 AppleScript 同步逻辑。

**Tech Stack:** Python, AppleScript (osascript), macOS Calendar.app.

---

### Task 1: 创建同步脚本

**Files:**
- Create: `CalendarAssistant/sync_v1_19_4.py`

- [ ] **Step 1: 编写同步脚本内容**

```python
import subprocess
from datetime import datetime

def sync_to_mac_calendar(event):
    s_day, s_month, s_year = event['start'].day, event['start'].strftime('%B'), event['start'].year
    s_hour, s_min = event['start'].hour, event['start'].minute
    e_day, e_month, e_year = event['end'].day, event['end'].strftime('%B'), event['end'].year
    e_hour, e_min = event['end'].hour, event['end'].minute

    script = f'''
    set startDate to (current date)
    set day of startDate to {s_day}
    set month of startDate to {s_month}
    set year of startDate to {s_year}
    set hours of startDate to {s_hour}
    set minutes of startDate to {s_min}
    set seconds of startDate to 0

    set endDate to (current date)
    set day of endDate to {e_day}
    set month of endDate to {e_month}
    set year of endDate to {e_year}
    set hours of endDate to {e_hour}
    set minutes of endDate to {e_min}
    set seconds of endDate to 0

    tell application "Calendar"
        try
            set theCal to first calendar whose name is "工作"
        on error
            set theCal to first calendar
        end try
        
        set currentEvent to make new event at theCal with properties {{summary:"{event['subject']}", start date:startDate, end date:endDate, location:"{event['link']}", description:"{event['description']}"}}
        
        tell currentEvent
            make new sound alarm at end with properties {{trigger interval:-15}}
            make new sound alarm at end with properties {{trigger interval:-5}}
            make new sound alarm at end with properties {{trigger interval:0}}
        end tell
        
        return "SUCCESS"
    end tell
    '''
    
    try:
        result = subprocess.check_output(["osascript", "-e", script]).decode('utf-8').strip()
        print(f"Result: {result}")
    except Exception as e:
        print(f"Failed: {e}")

if __name__ == "__main__":
    event = {
        "subject": "评审V1.19.4需求，【金山文档 | WPS云文档】 V1.19.4_禅道反馈需求优化",
        "start": datetime(2026, 4, 2, 10, 0),
        "end": datetime(2026, 4, 2, 11, 0),
        "link": "https://365.kdocs.cn/l/crrVk6GwcCV8",
        "description": "由辅助脚本自动创建"
    }
    sync_to_mac_calendar(event)
```

- [ ] **Step 2: 运行脚本并验证**

Run: `python3 CalendarAssistant/sync_v1_19_4.py`
Expected: 输出 `Result: SUCCESS`。同时请在 macOS 日历中确认日程。

- [ ] **Step 3: 提交 (可选)**

```bash
git add CalendarAssistant/sync_v1_19_4.py
git commit -m "feat: sync specific meeting for V1.19.4 review"
```
