--- +up
-- Create a generic trigger function to update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    -- Only update the timestamp if the row actually changed
    IF (OLD IS DISTINCT FROM NEW) THEN
        NEW.updated_at = now();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for all tables with updated_at column

-- profiles
CREATE TRIGGER update_profiles_updated_at
    BEFORE UPDATE ON public.profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- exercises
CREATE TRIGGER update_exercises_updated_at
    BEFORE UPDATE ON public.exercises
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- user_settings
CREATE TRIGGER update_user_settings_updated_at
    BEFORE UPDATE ON public.user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- routine_exercises
CREATE TRIGGER update_routine_exercises_updated_at
    BEFORE UPDATE ON public.routine_exercises
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- routine_sets
CREATE TRIGGER update_routine_sets_updated_at
    BEFORE UPDATE ON public.routine_sets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- routines
CREATE TRIGGER update_routines_updated_at
    BEFORE UPDATE ON public.routines
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- workouts
CREATE TRIGGER update_workouts_updated_at
    BEFORE UPDATE ON public.workouts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- workout_exercises
CREATE TRIGGER update_workout_exercises_updated_at
    BEFORE UPDATE ON public.workout_exercises
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- workout_images
CREATE TRIGGER update_workout_images_updated_at
    BEFORE UPDATE ON public.workout_images
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- workout_sets
CREATE TRIGGER update_workout_sets_updated_at
    BEFORE UPDATE ON public.workout_sets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- subscriptions
CREATE TRIGGER update_subscriptions_updated_at
    BEFORE UPDATE ON public.subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- user_devices
CREATE TRIGGER update_user_devices_updated_at
    BEFORE UPDATE ON public.user_devices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

--- +down
-- Drop triggers
DROP TRIGGER IF EXISTS update_profiles_updated_at ON public.profiles;
DROP TRIGGER IF EXISTS update_exercises_updated_at ON public.exercises;
DROP TRIGGER IF EXISTS update_user_settings_updated_at ON public.user_settings;
DROP TRIGGER IF EXISTS update_routine_exercises_updated_at ON public.routine_exercises;
DROP TRIGGER IF EXISTS update_routine_sets_updated_at ON public.routine_sets;
DROP TRIGGER IF EXISTS update_routines_updated_at ON public.routines;
DROP TRIGGER IF EXISTS update_workouts_updated_at ON public.workouts;
DROP TRIGGER IF EXISTS update_workout_exercises_updated_at ON public.workout_exercises;
DROP TRIGGER IF EXISTS update_workout_images_updated_at ON public.workout_images;
DROP TRIGGER IF EXISTS update_workout_sets_updated_at ON public.workout_sets;
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON public.subscriptions;
DROP TRIGGER IF EXISTS update_user_devices_updated_at ON public.user_devices;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();
