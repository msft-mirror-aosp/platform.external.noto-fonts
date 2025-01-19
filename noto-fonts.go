package noto_fonts
import (
  "android/soong/android"
  "fmt"
  prebuilt_etc "android/soong/etc"
  "os"
)

type fontProperties struct {
  // An optional font file of the versioned font files.
  // If not provided, the module name is used instead.
  FontFile string

  // An optional directory that contains versioned files.
  // If not provided, "font" is used instead.
  VersionDir string

  // A mandatory version flag used for controlling versioning.
  VersionFlag string

  // A mandatory version string used for environment that does not have build flags, e.g. SDK build.
  DefaultVersion string
}

type configProperties struct {
  // An optional font file of the versioned font files.
  // If not provided, "font_config.json" is used instead.
  ConfigFile string

  // An optional directory that contains versioned font files.
  // If not provided, "font" is used instead.
  VersionDir string

  // A mandatory version flag used for controlling versioning.
  VersionFlag string

  // A mandatory version string used for environment that does not have build flags, e.g. SDK build.
  DefaultVersion string
}

func init() {
  android.RegisterModuleType("prebuilt_versioned_font", prebuiltVersionedFontFactory)
  android.RegisterModuleType("versioned_font_config", versionedFontConfigFactory)
}

///////////////////////////////////////////////////////////////////////////////
// prebuilt_versioned_font definition
///////////////////////////////////////////////////////////////////////////////
func prebuiltVersionedFontFactory() (android.Module) {
  // Inherit prebuilt_font module
  module := prebuilt_etc.PrebuiltFontFactory()
  module.AddProperties(&fontProperties{})
  android.AddLoadHook(module, prebuiltVersionFont)
  return module
}

func prebuiltVersionFont(ctx android.LoadHookContext) {
  // Find fontProperties propery from the property list.
  var fontProp *fontProperties
  for _, p := range ctx.Module().GetProperties() {
      _fontProp, ok := p.(*fontProperties)
      if ok {
        fontProp = _fontProp
      }
  }

  // Check if the default version is specified.
  defaultVersion := fontProp.DefaultVersion
  if len(defaultVersion) == 0 {
    fmt.Printf("defaultVersion is required for SDK build")
    os.Exit(1)
  }

  // Read the version number from build flag.
  // The build flag may not be present, in such case, use default value instead.
  version, ok := ctx.Config().GetBuildFlag(fontProp.VersionFlag)
  if !ok {
    fmt.Printf("No value set to Build Flag %s. Use default version instead.", fontProp.VersionFlag)
    version = defaultVersion
  }

  // If no VersionDir is specified, use "font" as a default.
  if len(fontProp.VersionDir) == 0 {
    fontProp.VersionDir = "font"
  }

  // If no FontFile is specified, use module name as a default.
  if len(fontProp.FontFile) == 0 {
    fontProp.FontFile = ctx.Module().Name()
  }

  // Override the src property with newly generated one.
  type props struct {
    Src string
  }
  p := &props{}
  p.Src = fontProp.VersionDir + "/" + version + "/" + fontProp.FontFile
  ctx.AppendProperties(p)
}


///////////////////////////////////////////////////////////////////////////////
// versioned_font_config definition
///////////////////////////////////////////////////////////////////////////////
func versionedFontConfigFactory() (android.Module) {
  // Inherit filegroup module
  module := android.FileGroupFactory()
  module.AddProperties(&configProperties{})
  android.AddLoadHook(module, versionedFontConfig)
  return module
}

func versionedFontConfig(ctx android.LoadHookContext) {
  // Find fontProperties propery from the property list.
  var configProp *configProperties
  for _, p := range ctx.Module().GetProperties() {
      _configProp, ok := p.(*configProperties)
      if ok {
        configProp = _configProp
      }
  }

  // Check if the default version is specified.
  defaultVersion := configProp.DefaultVersion
  if len(defaultVersion) == 0 {
    fmt.Printf("defaultVersion is required for SDK build")
    os.Exit(1)
  }

  // Read the version number from build flag.
  // The build flag may not be present, in such case, use default value instead.
  version, ok := ctx.Config().GetBuildFlag(configProp.VersionFlag)
  if !ok {
    fmt.Printf("No value set to Build Flag %s. Use default version instead.", configProp.VersionFlag)
    version = defaultVersion
  }

  // If no VersionDir is specified, use "font" as a default.
  if len(configProp.VersionDir) == 0 {
    configProp.VersionDir = "font"
  }

  // If no ConfigFile is specified, use font_config.json instead
  if len(configProp.ConfigFile) == 0 {
    configProp.ConfigFile = "font_config.json"
  }

  // Override the src property with newly generated one.
  type props struct {
    Srcs []string
  }
  p := &props{}
  p.Srcs = make([]string, 1)
  p.Srcs[0] = configProp.VersionDir + "/" + version + "/" + configProp.ConfigFile
  ctx.AppendProperties(p)
}

