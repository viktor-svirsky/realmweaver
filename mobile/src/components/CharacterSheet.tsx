import React from 'react';
import { View, Text, Pressable, StyleSheet, ScrollView, Modal } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { getNameColor, getPKTitle } from '../types/game';

interface Props {
  visible: boolean;
  onClose: () => void;
}

export default function CharacterSheet({ visible, onClose }: Props) {
  const character = useGameStore((s) => s.character);

  if (!character) return null;

  return (
    <Modal visible={visible} animationType="slide" transparent>
      <View style={styles.overlay}>
        <View style={styles.sheet}>
          <View style={styles.header}>
            <Text style={styles.name}>{character.name}</Text>
            <Pressable
              onPress={onClose}
              style={styles.closeButtonContainer}
              accessibilityRole="button"
              accessibilityLabel="Close"
            >
              <Text style={styles.closeButton}>Close</Text>
            </Pressable>
          </View>

          <Text style={styles.classLevel}>
            Level {character.level} {character.class.charAt(0).toUpperCase() + character.class.slice(1)}
          </Text>

          <Text style={[styles.karmaText, { color: getNameColor(character.karma, character.flagged) }]}>
            {getPKTitle(character.karma, character.pk_count)}
          </Text>
          <Text style={styles.pkStats}>
            PK: {character.pk_count} | PvP: {character.pvp_count} | Karma: {character.karma}
          </Text>

          <View style={styles.barContainer}>
            <Text style={styles.barLabel}>HP</Text>
            <View style={styles.barBg}>
              <View style={[styles.barFill, styles.hpBar, { width: `${(character.hp / character.max_hp) * 100}%` }]} />
            </View>
            <Text style={styles.barValue}>{character.hp}/{character.max_hp}</Text>
          </View>

          <View style={styles.barContainer}>
            <Text style={styles.barLabel}>MP</Text>
            <View style={styles.barBg}>
              <View style={[styles.barFill, styles.manaBar, { width: `${(character.mana / character.max_mana) * 100}%` }]} />
            </View>
            <Text style={styles.barValue}>{character.mana}/{character.max_mana}</Text>
          </View>

          <View style={styles.barContainer}>
            <Text style={styles.barLabel}>XP</Text>
            <View style={styles.barBg}>
              <View style={[styles.barFill, styles.xpBar, { width: `${Math.min((character.xp / ((character.level + 1) ** 2 * 100)) * 100, 100)}%` }]} />
            </View>
            <Text style={styles.barValue}>{character.xp}</Text>
          </View>

          <ScrollView style={styles.statsSection}>
            <Text style={styles.sectionTitle}>Stats</Text>
            <View style={styles.statsGrid}>
              {Object.entries(character.stats).map(([key, value]) => (
                <View key={key} style={styles.statItem}>
                  <Text style={styles.statName}>{key.toUpperCase()}</Text>
                  <Text style={styles.statValue}>{value}</Text>
                  <Text style={styles.statMod}>
                    {Math.floor((value - 10) / 2) >= 0 ? '+' : ''}{Math.floor((value - 10) / 2)}
                  </Text>
                </View>
              ))}
            </View>

            <Text style={styles.sectionTitle}>Equipment</Text>
            <Text style={styles.equipItem}>
              Weapon: {character.equipment.weapon?.name || 'None'}
            </Text>
            <Text style={styles.equipItem}>
              Armor: {character.equipment.armor?.name || 'None'}
            </Text>
            <Text style={styles.equipItem}>AC: {character.ac}</Text>

            <Text style={styles.sectionTitle}>
              Inventory ({character.inventory.length} items)
            </Text>
            {character.inventory.map((item, i) => (
              <Text key={item.id || i} style={styles.invItem}>
                {item.name} {item.value > 0 ? `(${item.value}g)` : ''}
              </Text>
            ))}

            <Text style={styles.gold}>Gold: {character.gold}</Text>
          </ScrollView>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.7)',
    justifyContent: 'flex-end',
  },
  sheet: {
    backgroundColor: '#1a1a2e',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 20,
    maxHeight: '80%',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  name: {
    color: '#e0d68a',
    fontSize: 24,
    fontWeight: 'bold',
  },
  closeButtonContainer: {
    backgroundColor: '#2c3e50',
    paddingHorizontal: 20,
    paddingVertical: 8,
    borderRadius: 20,
  },
  closeButton: {
    color: '#e0d68a',
    fontSize: 14,
    fontWeight: 'bold',
  },
  classLevel: {
    color: '#a0a0b0',
    fontSize: 16,
    marginBottom: 4,
  },
  karmaText: {
    fontSize: 14,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  pkStats: {
    color: '#a0a0b0',
    fontSize: 12,
    marginBottom: 16,
  },
  barContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  barLabel: {
    color: '#a0a0b0',
    width: 30,
    fontSize: 12,
    fontWeight: 'bold',
  },
  barBg: {
    flex: 1,
    height: 12,
    backgroundColor: '#333',
    borderRadius: 6,
    overflow: 'hidden',
  },
  barFill: {
    height: '100%',
    borderRadius: 6,
  },
  hpBar: { backgroundColor: '#c0392b' },
  manaBar: { backgroundColor: '#2980b9' },
  xpBar: { backgroundColor: '#f39c12' },
  barValue: {
    color: '#a0a0b0',
    width: 60,
    textAlign: 'right',
    fontSize: 12,
  },
  statsSection: {
    marginTop: 16,
  },
  sectionTitle: {
    color: '#e0d68a',
    fontSize: 16,
    fontWeight: 'bold',
    marginTop: 12,
    marginBottom: 8,
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
  },
  statItem: {
    width: '33%',
    alignItems: 'center',
    marginBottom: 12,
  },
  statName: {
    color: '#a0a0b0',
    fontSize: 12,
  },
  statValue: {
    color: '#fff',
    fontSize: 22,
    fontWeight: 'bold',
  },
  statMod: {
    color: '#e0d68a',
    fontSize: 12,
  },
  equipItem: {
    color: '#d4d4d4',
    fontSize: 14,
    marginBottom: 4,
  },
  invItem: {
    color: '#a0a0b0',
    fontSize: 14,
    marginBottom: 2,
  },
  gold: {
    color: '#f39c12',
    fontSize: 16,
    fontWeight: 'bold',
    marginTop: 12,
  },
});
