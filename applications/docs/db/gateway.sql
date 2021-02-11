--
-- PostgreSQL database dump
--

-- Dumped from database version 12.4 (Debian 12.4-1.pgdg100+1)
-- Dumped by pg_dump version 12.4 (Debian 12.4-1.pgdg100+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_users; Type: TABLE; Schema: public; Owner: gateway
--

CREATE TABLE public.admin_users (
    uid text NOT NULL
);


ALTER TABLE public.admin_users OWNER TO gateway;

--
-- Name: applications; Type: TABLE; Schema: public; Owner: gateway
--

CREATE TABLE public.applications (
    application_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    description text,
    redirect_url text NOT NULL,
    application_name character varying(50),
    trim_app_name boolean DEFAULT false NOT NULL
);


ALTER TABLE public.applications OWNER TO gateway;

--
-- Name: admin_users admin_users_pkey; Type: CONSTRAINT; Schema: public; Owner: gateway
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_pkey PRIMARY KEY (uid);


--
-- Name: applications applications_pkey; Type: CONSTRAINT; Schema: public; Owner: gateway
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_pkey PRIMARY KEY (application_id);


--
-- PostgreSQL database dump complete
--
