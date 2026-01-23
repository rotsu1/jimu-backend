--- +up
-- Create a function to provision a new user on signup
CREATE OR REPLACE FUNCTION public.fn_on_signup_provisioning()
RETURNS TRIGGER AS $$
DECLARE
    new_profile_id uuid;
BEGIN
    -- 1. Create the Profile first (since identities needs profile_id)
    -- We use the returning ID to link everything together
    INSERT INTO public.profiles (name, primary_email)
    VALUES (
        COALESCE(NEW.provider_email, 'New Athlete'), -- Fallback if email isn't provided
        NEW.provider_email
    )
    RETURNING id INTO new_profile_id;

    -- 2. Create the User Settings for this new profile
    INSERT INTO public.user_settings (user_id, weight_unit, theme)
    VALUES (new_profile_id);

    -- 3. Update the incoming Identity record with the new Profile ID
    NEW.user_id := new_profile_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a trigger to add profile after inserting a new identity
CREATE TRIGGER tr_provision_user_on_signup
    BEFORE INSERT ON public.user_identities
    FOR EACH ROW
    WHEN (NEW.user_id IS NULL) -- Only run this if we haven't manually linked a user
    EXECUTE FUNCTION public.fn_on_signup_provisioning();

--- +down
DROP FUNCTION IF EXISTS public.fn_on_signup_provisioning;
DROP TRIGGER IF EXISTS tr_provision_user_on_signup ON public.user_identities;