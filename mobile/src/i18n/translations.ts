const translations: Record<string, Record<string, string>> = {
  en: {
    // UI buttons
    stats: 'Stats',
    map: 'Map',
    chat: 'Chat',
    close: 'Close',
    cancel: 'Cancel',
    travel: 'Travel',
    send: 'Send',
    // Home screen
    new_adventure: 'New Adventure',
    continue: 'Continue',
    language: 'Language',
    ai_dungeon_master: 'AI Dungeon Master',
    // Character creation
    character_name: 'Character Name',
    enter_name: 'Enter name...',
    choose_class: 'Choose Class',
    begin_adventure: 'Begin Adventure',
    creating: 'Creating...',
    // Classes
    class_warrior: 'Warrior',
    class_mage: 'Mage',
    class_rogue: 'Rogue',
    class_cleric: 'Cleric',
    class_ranger: 'Ranger',
    class_paladin: 'Paladin',
    class_necromancer: 'Necromancer',
    class_berserker: 'Berserker',
    class_warrior_desc: 'STR 16 CON 14 — Heavy armor, melee combat',
    class_mage_desc: 'INT 16 WIS 14 — Powerful spells, light armor',
    class_rogue_desc: 'DEX 16 INT 14 — Quick strikes, stealth',
    class_cleric_desc: 'WIS 16 CON 14 — Healing, divine magic',
    class_ranger_desc: 'DEX 16 WIS 14 — Ranged attacks, nature affinity',
    class_paladin_desc: 'STR 14 CHA 16 — Holy warrior, heals + fights',
    class_necromancer_desc: 'INT 16 CON 14 — Dark magic, life drain',
    class_berserker_desc: 'STR 18 CON 14 — Raw damage, rage mode',
    // Character sheet
    level: 'Level',
    equipment: 'Equipment',
    inventory: 'Inventory',
    gold: 'Gold',
    weapon: 'Weapon',
    armor: 'Armor',
    none: 'None',
    items: 'items',
    // Map
    world_map: 'World Map',
    current: 'Current',
    travel_to: 'Travel to',
    estimated_travel_time: 'Estimated travel time',
    tap_millhaven: 'Tap Millhaven to see locations',
    tap_center_hex: 'Tap your location to see details',
    tap_location_below: 'Tap a location below to explore',
    millhaven_locations: 'Millhaven Locations',
    people: 'People',
    nearby_players: 'Nearby Players',
    tap_player_interact: 'Tap a player to interact',
    no_players_nearby: 'No other players nearby',
    legend: 'Legend',
    explored: 'Explored',
    hinted: 'Hinted',
    unknown: 'Unknown',
    traveling: 'Traveling...',
    here: 'Here',
    // Chat
    region_chat: 'Region Chat',
    say_something: 'Say something...',
    no_messages: 'No messages yet. Say something!',
    // Combat
    round: 'Round',
    // Loading messages
    loading_1: 'The dungeon master is thinking...',
    loading_2: 'Rolling dice...',
    loading_3: 'Consulting ancient scrolls...',
    loading_4: 'Weaving the narrative...',
    loading_5: 'The world shifts around you...',
    loading_6: 'Fate is being decided...',
    // NPC roles
    role_innkeeper: 'Innkeeper',
    role_blacksmith: 'Blacksmith',
    role_mayor: 'Mayor',
    role_cleric: 'Cleric',
    role_merchant: 'Merchant',
    // Location descriptions
    danger: 'Danger!',
  },
  uk: {
    // UI buttons
    stats: '\u0421\u0442\u0430\u0442\u0438\u0441\u0442\u0438\u043A\u0430',
    map: '\u041A\u0430\u0440\u0442\u0430',
    chat: '\u0427\u0430\u0442',
    close: '\u0417\u0430\u043A\u0440\u0438\u0442\u0438',
    cancel: '\u0421\u043A\u0430\u0441\u0443\u0432\u0430\u0442\u0438',
    travel: '\u041F\u043E\u0434\u043E\u0440\u043E\u0436\u0443\u0432\u0430\u0442\u0438',
    send: '\u041D\u0430\u0434\u0456\u0441\u043B\u0430\u0442\u0438',
    // Home screen
    new_adventure: '\u041D\u043E\u0432\u0430 \u043F\u0440\u0438\u0433\u043E\u0434\u0430',
    continue: '\u041F\u0440\u043E\u0434\u043E\u0432\u0436\u0438\u0442\u0438',
    language: '\u041C\u043E\u0432\u0430',
    ai_dungeon_master: '\u0428\u0442\u0443\u0447\u043D\u0438\u0439 \u043C\u0430\u0439\u0441\u0442\u0435\u0440 \u043F\u0456\u0434\u0437\u0435\u043C\u0435\u043B\u044C',
    // Character creation
    character_name: "\u0406\u043C'\u044F \u043F\u0435\u0440\u0441\u043E\u043D\u0430\u0436\u0430",
    enter_name: "\u0412\u0432\u0435\u0434\u0456\u0442\u044C \u0456\u043C'\u044F...",
    choose_class: '\u041E\u0431\u0435\u0440\u0456\u0442\u044C \u043A\u043B\u0430\u0441',
    begin_adventure: '\u041F\u043E\u0447\u0430\u0442\u0438 \u043F\u0440\u0438\u0433\u043E\u0434\u0443',
    creating: '\u0421\u0442\u0432\u043E\u0440\u0435\u043D\u043D\u044F...',
    // Classes
    class_warrior: '\u0412\u043E\u0457\u043D',
    class_mage: '\u041C\u0430\u0433',
    class_rogue: '\u0420\u043E\u0437\u0431\u0456\u0439\u043D\u0438\u043A',
    class_cleric: '\u041A\u043B\u0435\u0440\u0438\u043A',
    class_ranger: '\u0420\u0435\u0439\u043D\u0434\u0436\u0435\u0440',
    class_paladin: '\u041F\u0430\u043B\u0430\u0434\u0456\u043D',
    class_necromancer: '\u041D\u0435\u043A\u0440\u043E\u043C\u0430\u043D\u0442',
    class_berserker: '\u0411\u0435\u0440\u0441\u0435\u0440\u043A',
    class_warrior_desc: 'STR 16 CON 14 \u2014 \u0412\u0430\u0436\u043A\u0430 \u0431\u0440\u043E\u043D\u044F, \u0431\u043B\u0438\u0436\u043D\u0456\u0439 \u0431\u0456\u0439',
    class_mage_desc: 'INT 16 WIS 14 \u2014 \u041F\u043E\u0442\u0443\u0436\u043D\u0456 \u0437\u0430\u043A\u043B\u0438\u043D\u0430\u043D\u043D\u044F, \u043B\u0435\u0433\u043A\u0430 \u0431\u0440\u043E\u043D\u044F',
    class_rogue_desc: 'DEX 16 INT 14 \u2014 \u0428\u0432\u0438\u0434\u043A\u0456 \u0443\u0434\u0430\u0440\u0438, \u0441\u043A\u0440\u0438\u0442\u043D\u0456\u0441\u0442\u044C',
    class_cleric_desc: 'WIS 16 CON 14 \u2014 \u0417\u0446\u0456\u043B\u0435\u043D\u043D\u044F, \u0431\u043E\u0436\u0435\u0441\u0442\u0432\u0435\u043D\u043D\u0430 \u043C\u0430\u0433\u0456\u044F',
    class_ranger_desc: 'DEX 16 WIS 14 \u2014 \u0414\u0430\u043B\u044C\u043D\u0456 \u0430\u0442\u0430\u043A\u0438, \u0441\u043F\u043E\u0440\u0456\u0434\u043D\u0435\u043D\u0456\u0441\u0442\u044C \u0437 \u043F\u0440\u0438\u0440\u043E\u0434\u043E\u044E',
    class_paladin_desc: 'STR 14 CHA 16 \u2014 \u0421\u0432\u044F\u0442\u0438\u0439 \u0432\u043E\u0457\u043D, \u0437\u0446\u0456\u043B\u044E\u0454 \u0439 \u0431\u2019\u0454\u0442\u044C\u0441\u044F',
    class_necromancer_desc: 'INT 16 CON 14 \u2014 \u0422\u0435\u043C\u043D\u0430 \u043C\u0430\u0433\u0456\u044F, \u0432\u0438\u0441\u043C\u043E\u043A\u0442\u0443\u0432\u0430\u043D\u043D\u044F \u0436\u0438\u0442\u0442\u044F',
    class_berserker_desc: 'STR 18 CON 14 \u2014 \u0413\u0440\u0443\u0431\u0430 \u0441\u0438\u043B\u0430, \u0440\u0435\u0436\u0438\u043C \u043B\u044E\u0442\u0456',
    // Character sheet
    level: '\u0420\u0456\u0432\u0435\u043D\u044C',
    equipment: '\u0421\u043F\u043E\u0440\u044F\u0434\u0436\u0435\u043D\u043D\u044F',
    inventory: '\u0406\u043D\u0432\u0435\u043D\u0442\u0430\u0440',
    gold: '\u0417\u043E\u043B\u043E\u0442\u043E',
    weapon: '\u0417\u0431\u0440\u043E\u044F',
    armor: '\u0411\u0440\u043E\u043D\u044F',
    none: '\u041D\u0435\u043C\u0430\u0454',
    items: '\u043F\u0440\u0435\u0434\u043C\u0435\u0442\u0456\u0432',
    // Map
    world_map: '\u041A\u0430\u0440\u0442\u0430 \u0441\u0432\u0456\u0442\u0443',
    current: '\u041F\u043E\u0442\u043E\u0447\u043D\u0430',
    travel_to: '\u041F\u043E\u0434\u043E\u0440\u043E\u0436\u0443\u0432\u0430\u0442\u0438 \u0434\u043E',
    estimated_travel_time: '\u041E\u0440\u0456\u0454\u043D\u0442\u043E\u0432\u043D\u0438\u0439 \u0447\u0430\u0441 \u043F\u043E\u0434\u043E\u0440\u043E\u0436\u0456',
    tap_millhaven: '\u041D\u0430\u0442\u0438\u0441\u043D\u0456\u0442\u044C \u043D\u0430 Millhaven, \u0449\u043E\u0431 \u043F\u043E\u0431\u0430\u0447\u0438\u0442\u0438 \u043B\u043E\u043A\u0430\u0446\u0456\u0457',
    tap_center_hex: '\u041D\u0430\u0442\u0438\u0441\u043D\u0456\u0442\u044C \u043D\u0430 \u0432\u0430\u0448\u0443 \u043B\u043E\u043A\u0430\u0446\u0456\u044E \u0434\u043B\u044F \u0434\u0435\u0442\u0430\u043B\u0435\u0439',
    tap_location_below: '\u041D\u0430\u0442\u0438\u0441\u043D\u0456\u0442\u044C \u043D\u0430 \u043B\u043E\u043A\u0430\u0446\u0456\u044E \u043D\u0438\u0436\u0447\u0435',
    millhaven_locations: '\u041B\u043E\u043A\u0430\u0446\u0456\u0457 Millhaven',
    people: '\u041B\u044E\u0434\u0438',
    nearby_players: '\u0413\u0440\u0430\u0432\u0446\u0456 \u043F\u043E\u0431\u043B\u0438\u0437\u0443',
    tap_player_interact: '\u041D\u0430\u0442\u0438\u0441\u043D\u0456\u0442\u044C \u043D\u0430 \u0433\u0440\u0430\u0432\u0446\u044F \u0434\u043B\u044F \u0432\u0437\u0430\u0454\u043C\u043E\u0434\u0456\u0457',
    no_players_nearby: '\u041D\u0435\u043C\u0430\u0454 \u0456\u043D\u0448\u0438\u0445 \u0433\u0440\u0430\u0432\u0446\u0456\u0432 \u043F\u043E\u0431\u043B\u0438\u0437\u0443',
    legend: '\u041B\u0435\u0433\u0435\u043D\u0434\u0430',
    explored: '\u0414\u043E\u0441\u043B\u0456\u0434\u0436\u0435\u043D\u043E',
    hinted: '\u041D\u0430\u0442\u044F\u043A\u043D\u0443\u0442\u043E',
    unknown: '\u041D\u0435\u0432\u0456\u0434\u043E\u043C\u043E',
    traveling: '\u041F\u043E\u0434\u043E\u0440\u043E\u0436\u0443\u0454\u043C\u043E...',
    here: '\u0422\u0443\u0442',
    // Chat
    region_chat: '\u0427\u0430\u0442 \u0440\u0435\u0433\u0456\u043E\u043D\u0443',
    say_something: '\u041D\u0430\u043F\u0438\u0448\u0456\u0442\u044C \u0449\u043E\u0441\u044C...',
    no_messages: '\u041F\u043E\u043A\u0438 \u043D\u0435\u043C\u0430\u0454 \u043F\u043E\u0432\u0456\u0434\u043E\u043C\u043B\u0435\u043D\u044C. \u041D\u0430\u043F\u0438\u0448\u0456\u0442\u044C \u0449\u043E\u0441\u044C!',
    // Combat
    round: '\u0420\u0430\u0443\u043D\u0434',
    // Loading messages
    loading_1: '\u041C\u0430\u0439\u0441\u0442\u0435\u0440 \u043F\u0456\u0434\u0437\u0435\u043C\u0435\u043B\u044C \u0434\u0443\u043C\u0430\u0454...',
    loading_2: '\u041A\u0438\u0434\u0430\u0454\u043C\u043E \u043A\u0443\u0431\u0438\u043A\u0438...',
    loading_3: '\u0420\u0430\u0434\u0438\u043C\u043E\u0441\u044F \u0437 \u0441\u0442\u0430\u0440\u043E\u0434\u0430\u0432\u043D\u0456\u043C\u0438 \u0441\u0443\u0432\u043E\u044F\u043C\u0438...',
    loading_4: '\u041F\u043B\u0435\u0442\u0435\u043C\u043E \u043E\u043F\u043E\u0432\u0456\u0434\u044C...',
    loading_5: '\u0421\u0432\u0456\u0442 \u0437\u043C\u0456\u043D\u044E\u0454\u0442\u044C\u0441\u044F \u043D\u0430\u0432\u043A\u043E\u043B\u043E \u0432\u0430\u0441...',
    loading_6: '\u0414\u043E\u043B\u044F \u0432\u0438\u0440\u0456\u0448\u0443\u0454\u0442\u044C\u0441\u044F...',
    // NPC roles
    role_innkeeper: '\u0428\u0438\u043D\u043A\u0430\u0440\u043A\u0430',
    role_blacksmith: '\u041A\u043E\u0432\u0430\u043B\u044C',
    role_mayor: '\u0421\u0442\u0430\u0440\u043E\u0441\u0442\u0430',
    role_cleric: '\u041A\u043B\u0435\u0440\u0438\u043A',
    role_merchant: '\u0422\u043E\u0440\u0433\u043E\u0432\u0435\u0446\u044C',
    // Location descriptions
    danger: '\u041D\u0435\u0431\u0435\u0437\u043F\u0435\u043A\u0430!',
  },
};

export function t(key: string, lang: string): string {
  return translations[lang]?.[key] || translations.en?.[key] || key;
}

/**
 * Translate known action labels that come from the server.
 * NPC/location proper names stay untranslated.
 */
export function translateLabel(label: string, lang: string): string {
  if (lang === 'en') return label;

  const labelMap: Record<string, Record<string, string>> = {
    uk: {
      'Look Around': '\u041E\u0433\u043B\u044F\u043D\u0443\u0442\u0438\u0441\u044F',
      'Travel': '\u041F\u043E\u0434\u043E\u0440\u043E\u0436\u0443\u0432\u0430\u0442\u0438',
      'Rest': '\u0412\u0456\u0434\u043F\u043E\u0447\u0438\u0442\u0438',
      'Attack': '\u0410\u0442\u0430\u043A\u0443\u0432\u0430\u0442\u0438',
      'Defend': '\u0417\u0430\u0445\u0438\u0449\u0430\u0442\u0438\u0441\u044F',
      'Flee': '\u0412\u0442\u0435\u043A\u0442\u0438',
      'Use Item': '\u0412\u0438\u043A\u043E\u0440\u0438\u0441\u0442\u0430\u0442\u0438 \u043F\u0440\u0435\u0434\u043C\u0435\u0442',
      'Ask about quests': '\u0417\u0430\u043F\u0438\u0442\u0430\u0442\u0438 \u043F\u0440\u043E \u0437\u0430\u0432\u0434\u0430\u043D\u043D\u044F',
      'Ask for rumors': '\u0417\u0430\u043F\u0438\u0442\u0430\u0442\u0438 \u043F\u0440\u043E \u0447\u0443\u0442\u043A\u0438',
      'Ask about the cave': '\u0417\u0430\u043F\u0438\u0442\u0430\u0442\u0438 \u043F\u0440\u043E \u043F\u0435\u0447\u0435\u0440\u0443',
      'Trade': '\u0422\u043E\u0440\u0433\u0443\u0432\u0430\u0442\u0438',
      'Leave conversation': '\u0417\u0430\u043A\u0456\u043D\u0447\u0438\u0442\u0438 \u0440\u043E\u0437\u043C\u043E\u0432\u0443',
      'Hunt Enemies': '\u041F\u043E\u043B\u044E\u0432\u0430\u043D\u043D\u044F',
    },
  };

  // Exact match
  if (labelMap[lang]?.[label]) {
    return labelMap[lang][label];
  }

  // Prefix pattern: "Talk to X" -> "Поговорити з X"
  const prefixMap: Record<string, Record<string, string>> = {
    uk: {
      'Talk to ': '\u041F\u043E\u0433\u043E\u0432\u043E\u0440\u0438\u0442\u0438 \u0437 ',
      'Enter ': '\u0423\u0432\u0456\u0439\u0442\u0438 \u0432 ',
      'Go to ': '\u041F\u0456\u0442\u0438 \u0434\u043E ',
      'Explore ': '\u0414\u043E\u0441\u043B\u0456\u0434\u0438\u0442\u0438 ',
    },
  };

  if (prefixMap[lang]) {
    for (const [prefix, translated] of Object.entries(prefixMap[lang])) {
      if (label.startsWith(prefix)) {
        return translated + label.slice(prefix.length);
      }
    }
  }

  return label;
}
