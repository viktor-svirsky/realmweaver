import React, { useEffect, useState } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { StatusBar } from 'expo-status-bar';
import { View, Text, Pressable, StyleSheet } from 'react-native';
import { authenticate } from './src/api/nakama';
import { connectSocket } from './src/api/socket';
import HomeScreen from './src/screens/HomeScreen';
import CharacterCreateScreen from './src/screens/CharacterCreateScreen';
import GameScreen from './src/screens/GameScreen';

export type RootStackParamList = {
  Home: undefined;
  CharacterCreate: undefined;
  Game: { characterId: string };
};

const Stack = createNativeStackNavigator<RootStackParamList>();

export default function App() {
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function init() {
    setError(null);
    try {
      await authenticate();
      await connectSocket();
      setReady(true);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Unknown error';
      console.error('Init failed:', msg);
      setError(msg);
    }
  }

  useEffect(() => {
    init();
  }, []);

  if (!ready) {
    return (
      <View style={styles.loading}>
        <Text style={styles.loadingTitle}>Realmweaver</Text>
        {error ? (
          <>
            <Text style={styles.errorText}>Connection failed: {error}</Text>
            <Pressable style={styles.retryButton} onPress={init}>
              <Text style={styles.retryText}>Retry</Text>
            </Pressable>
          </>
        ) : (
          <Text style={styles.loadingText}>Connecting...</Text>
        )}
      </View>
    );
  }

  return (
    <NavigationContainer>
      <StatusBar style="light" />
      <Stack.Navigator
        screenOptions={{
          headerStyle: { backgroundColor: '#1a1a2e' },
          headerTintColor: '#e0d68a',
          headerTitleStyle: { fontWeight: 'bold' },
          contentStyle: { backgroundColor: '#16213e' },
        }}
      >
        <Stack.Screen
          name="Home"
          component={HomeScreen}
          options={{ title: 'Realmweaver' }}
        />
        <Stack.Screen
          name="CharacterCreate"
          component={CharacterCreateScreen}
          options={{ title: 'Create Character' }}
        />
        <Stack.Screen
          name="Game"
          component={GameScreen}
          options={{ title: 'Adventure', headerShown: false }}
        />
      </Stack.Navigator>
    </NavigationContainer>
  );
}

const styles = StyleSheet.create({
  loading: {
    flex: 1,
    backgroundColor: '#16213e',
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingTitle: {
    color: '#e0d68a',
    fontSize: 32,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  loadingText: {
    color: '#a0a0b0',
    fontSize: 16,
  },
  errorText: {
    color: '#c0392b',
    fontSize: 14,
    marginBottom: 16,
    textAlign: 'center',
    paddingHorizontal: 40,
  },
  retryButton: {
    backgroundColor: '#e0d68a',
    paddingHorizontal: 32,
    paddingVertical: 12,
    borderRadius: 8,
  },
  retryText: {
    color: '#1a1a2e',
    fontSize: 16,
    fontWeight: 'bold',
  },
});
