// 中文标点符号纠错工具
// 基于 yiwo-insights-studio/.claude/skills/write-article/bin/fix_punctuation.py 的算法
// 仅在复制时使用，不修改原稿件。

const CJK = "[一-鿿]";

function replaceQuotes(text) {
  if (!text) return text;
  let result = "";
  let isOpen = false;
  let i = 0;
  while (i < text.length) {
    let ch = text[i];
    // 处理转义的 \"
    if (ch === "\\" && i + 1 < text.length && text[i + 1] === '"') {
      i++;
      ch = '"';
    }
    if (ch === '"') {
      result += isOpen ? "”" : "“";
      isOpen = !isOpen;
    } else {
      result += ch;
    }
    i++;
  }
  return result;
}

function replacePunctuation(text) {
  if (!text) return text;

  // 感叹号 ! -> ！
  text = text.replace(new RegExp(`(?<=${CJK})!`, "g"), "！");
  text = text.replace(new RegExp(`!\\s*(?=${CJK})`, "g"), "！");

  // 问号 ? -> ？
  text = text.replace(new RegExp(`(?<=${CJK})\\?`, "g"), "？");
  text = text.replace(new RegExp(`\\?\\s*(?=${CJK})`, "g"), "？");

  // 逗号 , -> ，
  text = text.replace(new RegExp(`(?<=${CJK}),`, "g"), "，");
  text = text.replace(new RegExp(`,\\s*(?=${CJK})`, "g"), "，");

  // 冒号 : -> ：
  text = text.replace(new RegExp(`(?<=${CJK}):`, "g"), "：");
  text = text.replace(new RegExp(`:\\s*(?=${CJK})`, "g"), "：");

  // 分号 ; -> ；
  text = text.replace(new RegExp(`(?<=${CJK});`, "g"), "；");
  text = text.replace(new RegExp(`;\\s*(?=${CJK})`, "g"), "；");

  return text;
}

export function fixPunctuation(text) {
  if (!text) return text;
  return replacePunctuation(replaceQuotes(text));
}
