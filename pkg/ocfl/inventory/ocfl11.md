<!DOCTYPE html>
<!-- saved from url=(0025)https://ocfl.io/1.1/spec/ -->
<html lang="en-US"><head><meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">

<!-- Begin Jekyll SEO tag v2.8.0 -->
<title>Oxford Common File Layout | This Oxford Common File Layout (OCFL) specification describes an application-independent approach to the storage of digital information in a structured, transparent, and predictable manner.</title>
<meta name="generator" content="Jekyll v3.10.0">
<meta property="og:title" content="Oxford Common File Layout">
<meta property="og:locale" content="en_US">
<meta name="description" content="This Oxford Common File Layout (OCFL) specification describes an application-independent approach to the storage of digital information in a structured, transparent, and predictable manner.">
<meta property="og:description" content="This Oxford Common File Layout (OCFL) specification describes an application-independent approach to the storage of digital information in a structured, transparent, and predictable manner.">
<link rel="canonical" href="https://ocfl.io/1.1/spec/">
<meta property="og:url" content="https://ocfl.io/1.1/spec/">
<meta property="og:site_name" content="Oxford Common File Layout">
<meta property="og:type" content="website">
<meta name="twitter:card" content="summary">
<meta property="twitter:title" content="Oxford Common File Layout">
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"WebPage","description":"This Oxford Common File Layout (OCFL) specification describes an application-independent approach to the storage of digital information in a structured, transparent, and predictable manner.","headline":"Oxford Common File Layout","url":"https://ocfl.io/1.1/spec/"}</script>
<!-- End Jekyll SEO tag -->

    <style class="anchorjs"></style><link rel="stylesheet" href="./ocfl11_files/style.css">
    <style>
            .rfc2119 {
                text-transform: lowercase;
                font-variant: small-caps;
                font-style: normal;
                color: #900;
            }
    </style>
    <!-- start custom head snippets, customize with your own _includes/head-custom.html file -->

<!-- Setup Google Analytics -->



<!-- You can set your favicon here -->
<!-- link rel="shortcut icon" type="image/x-icon" href="/favicon.ico" -->

<!-- end custom head snippets -->

  <style id="gridman-extension-fonts">
  @font-face {
    font-display: auto;
    font-family: "SF Pro Display";
    font-style: normal;
    font-weight: 500;
    src: url("chrome-extension://cmplbmppmfboedgkkelpkfgaakabpicn/fonts/SF-Pro-Display-Regular.ttf") format("truetype");
  }
</style></head>
  <body cz-shortcut-listen="true">
    <!-- based on https://github.com/pages-themes/primer/blob/master/_layouts/default.html -->
    <div class="container-lg px-3 my-5 markdown-body">
      

      <p><img src="./ocfl11_files/35607965" alt="OCFL Hand-drive logo" style="float:right;width:307px;height:307px;"></p>
<h1 class="no_toc" id="oxford-common-file-layout-specification">Oxford Common File Layout Specification</h1>

<p>7 October 2022, updated 7 November 2024</p>

<p><strong>This Version:</strong></p>
<ul>
  <li><a href="https://ocfl.io/1.1/spec/">https://ocfl.io/1.1/spec/</a></li>
</ul>

<p><strong>Latest Published Version:</strong></p>
<ul>
  <li><a href="https://ocfl.io/latest/spec/">https://ocfl.io/latest/spec/</a></li>
</ul>

<p><strong>Editors:</strong></p>

<ul>
  <li><a href="https://orcid.org/0000-0003-3311-3741">Neil Jefferies</a>, <br>
<a href="http://www.bodleian.ox.ac.uk/">Bodleian Libraries, University of Oxford</a></li>
  <li><a href="https://orcid.org/0000-0003-3526-2230">Rosalyn Metz</a>, <a href="https://web.library.emory.edu/">Emory University</a></li>
  <li><a href="https://orcid.org/0000-0003-4176-1933">Julian Morley</a>, <a href="https://library.stanford.edu/">Stanford University</a></li>
  <li><a href="https://orcid.org/0000-0002-7970-7855">Simeon Warner</a>, <a href="https://www.library.cornell.edu/">Cornell University</a></li>
  <li><a href="https://orcid.org/0000-0002-8318-4225">Andrew Woods</a>, <a href="https://library.harvard.edu/">Harvard University</a></li>
</ul>

<p><strong>Former Editors:</strong></p>

<ul>
  <li><a href="https://orcid.org/0000-0003-2663-0003">Andrew Hankinson</a></li>
</ul>

<p><strong>Additional Documents:</strong></p>

<ul>
  <li><a href="https://ocfl.io/1.1/implementation-notes/">Implementation Notes</a></li>
  <li><a href="https://ocfl.io/1.1/spec/change-log.html">Specification Change Log</a></li>
  <li><a href="https://ocfl.io/1.1/spec/validation-codes.html">Validation Codes</a></li>
  <li><a href="https://github.com/OCFL/extensions/">Extensions</a></li>
</ul>

<p><strong>Previous Version:</strong></p>
<ul>
  <li><a href="https://ocfl.io/1.1/spec/">https://ocfl.io/1.1/spec/</a></li>
</ul>

<p><strong>Repository:</strong></p>
<ul>
  <li><a href="https://github.com/ocfl/spec">Github</a></li>
  <li><a href="https://github.com/ocfl/spec/issues">Issues</a></li>
  <li><a href="https://github.com/ocfl/spec/commits">Commits</a></li>
  <li><a href="https://github.com/ocfl/Use-Cases">Use Cases</a></li>
</ul>

<p>This document is licensed under a <a href="https://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0
License</a>. <a href="./ocfl11_files/35607965">OCFL logo:
“hand-drive”</a> by
<a href="http://orcid.org/0000-0001-8390-6171">Patrick Hochstenbach</a> is
licensed under <a href="https://creativecommons.org/licenses/by/2.0/">CC BY 2.0</a>.</p>

<h2 class="no_toc" id="abstract">Introduction<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#abstract" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<p><em>This section is non-normative.</em></p>

<p>This Oxford Common File Layout (OCFL) specification describes an application-independent approach to the storage of
digital objects in a structured, transparent, and predictable manner. It is designed to promote long-term access and
management of digital objects within digital repositories.</p>

<h3 class="no_toc" id="need">Need<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#need" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The OCFL initiative began as a discussion amongst digital repository practitioners to identify well-defined, common, and
application-independent file management for a digital repository’s persisted objects and represents a specification of
the community’s collective recommendations addressing five primary requirements: completeness, parsability,
versioning, robustness, and storage diversity.</p>

<h4 class="no_toc" id="completeness">Completeness<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#completeness" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>The OCFL recommends storing metadata and the content it describes together so the OCFL object can be fully understood in
the absence of original software. The OCFL does not make recommendations about what constitutes an object, nor does it
assume what type of metadata is needed to fully understand the object, recognizing those decisions may differ from one
repository to another. However, it is recommended that when making this decision, implementers consider what is
necessary to rebuild the objects from the files stored.</p>

<h4 class="no_toc" id="parsability">Parsability<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#parsability" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>One goal of the OCFL is to ensure objects remain fixed over time. This can be difficult as software and infrastructure
change, and content is migrated. To combat this challenge, the OCFL ensures that both humans and machines can understand
the layout and corresponding inventory regardless of the software or infrastructure used. This allows for humans to read
the layout and corresponding inventory, and understand it without the use of machines. Additionally, if existing
software were to become obsolete, the OCFL could easily be understood by a light weight application, even without the
full feature repository that might have been used in the past.</p>

<h4 class="no_toc" id="versioning">Versioning<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#versioning" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>Another need expressed by the community was the need to update and change objects, either the content itself or the
metadata associated with the object. The OCFL relies heavily on the prior art in the [<a href="https://ocfl.io/1.1/spec/#ref-moab">Moab</a>] Design for
Digital Object Versioning which utilizes forward deltas to track the history of the object. Utilizing this schema allows
implementers of the OCFL to easily recreate past versions of an OCFL object. Like with objects, the OCFL remains silent
on when versioning should occur recognizing this may differ from implementation to implementation.</p>

<h4 class="no_toc" id="robustness">Robustness<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#robustness" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>The OCFL also fills the need for robustness against errors, corruption, and migration. The versioning schema ensures an
OCFL object is robust enough to allow for the discovery of human errors. The fixity checking built into the OCFL via
content addressable storage allows implementers to identify file corruption that might happen outside of normal human
interactions. The OCFL eases content migrations by providing a technology agnostic method for verifying OCFL objects
have remained fixed.</p>

<h4 class="no_toc" id="storage-diversity">Storage diversity<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#storage-diversity" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>Finally, the community expressed a need to store content on a wide variety of storage technologies. With that in mind,
the OCFL was written with an eye toward various storage infrastructures including cloud object stores.</p>

<h3 class="no_toc" id="note">Note<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#note" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>This normative specification describes the nature of an OCFL Object (the “object-at-rest”) and the arrangement of OCFL
Objects under an OCFL Storage Root. A set of recommendations for how OCFL Objects should be acted upon (the
“object-in-motion”) can be found in the [<a href="https://ocfl.io/1.1/spec/#ref-ocfl-implementation-notes">OCFL-Implementation-Notes</a>]. The OCFL
editorial group recommends reading both the specification and the implementation notes in order to understand the full
scope of the OCFL.</p>

<p>This specification is designed to operate on storage systems that employ a hierarchical metaphor for presenting data to
users. On traditional disk-based storage this may take the form of files and directories, and this is the terminology we
use in this specification since it is widely known. However, it may equally apply to object stores, where namespaces,
containers, and objects present a similar organization hierarchy to users.</p>

<h2 class="no_toc" id="table-of-contents">Table of Contents<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#table-of-contents" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<ul id="markdown-toc">
  <li><a href="https://ocfl.io/1.1/spec/#conformance" id="markdown-toc-conformance">1. Conformance</a></li>
  <li><a href="https://ocfl.io/1.1/spec/#terminology" id="markdown-toc-terminology">2. Terminology</a></li>
  <li><a href="https://ocfl.io/1.1/spec/#object-spec" id="markdown-toc-object-spec">3. OCFL Object</a>    <ul>
      <li><a href="https://ocfl.io/1.1/spec/#object-structure" id="markdown-toc-object-structure">3.1 Object Structure</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#object-conformance-declaration" id="markdown-toc-object-conformance-declaration">3.2 Object Conformance Declaration</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#version-directories" id="markdown-toc-version-directories">3.3 Version Directories</a>        <ul>
          <li><a href="https://ocfl.io/1.1/spec/#content-directory" id="markdown-toc-content-directory">3.3.1 Content Directory</a></li>
        </ul>
      </li>
      <li><a href="https://ocfl.io/1.1/spec/#digests" id="markdown-toc-digests">3.4 Digests</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#inventory" id="markdown-toc-inventory">3.5 Inventory</a>        <ul>
          <li><a href="https://ocfl.io/1.1/spec/#inventory-structure" id="markdown-toc-inventory-structure">3.5.1 Basic Structure</a></li>
          <li><a href="https://ocfl.io/1.1/spec/#manifest" id="markdown-toc-manifest">3.5.2 Manifest</a></li>
          <li><a href="https://ocfl.io/1.1/spec/#versions" id="markdown-toc-versions">3.5.3 Versions</a>            <ul>
              <li><a href="https://ocfl.io/1.1/spec/#version" id="markdown-toc-version">3.5.3.1 Version</a></li>
            </ul>
          </li>
          <li><a href="https://ocfl.io/1.1/spec/#fixity" id="markdown-toc-fixity">3.5.4 Fixity</a></li>
        </ul>
      </li>
      <li><a href="https://ocfl.io/1.1/spec/#inventory-digest" id="markdown-toc-inventory-digest">3.6 Inventory Digest</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#version-inventory" id="markdown-toc-version-inventory">3.7 Version Inventory and Inventory Digest</a>        <ul>
          <li><a href="https://ocfl.io/1.1/spec/#conformance-of-prior-versions" id="markdown-toc-conformance-of-prior-versions">3.7.1 Conformance of prior versions</a></li>
        </ul>
      </li>
      <li><a href="https://ocfl.io/1.1/spec/#logs-directory" id="markdown-toc-logs-directory">3.8 Logs Directory</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#object-extensions" id="markdown-toc-object-extensions">3.9 Object Extensions</a></li>
    </ul>
  </li>
  <li><a href="https://ocfl.io/1.1/spec/#storage-root" id="markdown-toc-storage-root">4. OCFL Storage Root</a>    <ul>
      <li><a href="https://ocfl.io/1.1/spec/#root-structure" id="markdown-toc-root-structure">4.1 Root Structure</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#root-conformance-declaration" id="markdown-toc-root-conformance-declaration">4.2 Root Conformance Declaration</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#root-hierarchies" id="markdown-toc-root-hierarchies">4.3 Storage Hierarchies</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#storage-root-extensions" id="markdown-toc-storage-root-extensions">4.4 Storage Root Extensions</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#documenting-local-extensions" id="markdown-toc-documenting-local-extensions">4.5 Documenting Local Extensions</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#filesystem-features" id="markdown-toc-filesystem-features">4.6 Filesystem features</a></li>
    </ul>
  </li>
  <li><a href="https://ocfl.io/1.1/spec/#examples" id="markdown-toc-examples">5. Examples</a>    <ul>
      <li><a href="https://ocfl.io/1.1/spec/#example-minimal-object" id="markdown-toc-example-minimal-object">5.1 Minimal OCFL Object</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#example-versioned-object" id="markdown-toc-example-versioned-object">5.2 Versioned OCFL Object</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#example-object-diff-paths" id="markdown-toc-example-object-diff-paths">5.3 Different Logical and Content Paths in an OCFL Object</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#example-bagit-in-ocfl" id="markdown-toc-example-bagit-in-ocfl">5.4 BagIt in an OCFL Object</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#example-moab-in-ocfl" id="markdown-toc-example-moab-in-ocfl">5.5 Moab in an OCFL Object</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#example-extended-storage-root" id="markdown-toc-example-extended-storage-root">5.6 Example Extended OCFL Storage Root</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#example-extended-object" id="markdown-toc-example-extended-object">5.7 Example Extended OCFL Object</a></li>
    </ul>
  </li>
  <li><a href="https://ocfl.io/1.1/spec/#references" id="markdown-toc-references">6. References</a>    <ul>
      <li><a href="https://ocfl.io/1.1/spec/#normative-references" id="markdown-toc-normative-references">6.1 Normative References</a></li>
      <li><a href="https://ocfl.io/1.1/spec/#informative-references" id="markdown-toc-informative-references">6.2 Informative References</a></li>
    </ul>
  </li>
  <li><a href="https://ocfl.io/1.1/spec/#revision-history" id="markdown-toc-revision-history">Revision history</a></li>
</ul>

<h2 id="conformance">1. Conformance<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#conformance" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<p>As well as sections marked as non-normative, all authoring guidelines, diagrams, examples, and notes in this
specification are non-normative. Everything else in this specification is normative.</p>

<p>The key words <span class="rfc2119">MAY</span>, <span class="rfc2119">MUST</span>, <span class="rfc2119">MUST
NOT</span>, <span class="rfc2119">SHOULD</span>, and <span class="rfc2119">SHOULD NOT</span> are to be interpreted as
described in [<a href="https://ocfl.io/1.1/spec/#ref-rfc2119">RFC2119</a>].</p>

<h2 id="terminology">2. Terminology<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#terminology" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<ul>
  <li>
    <p><a name="dfn-content-path"></a><strong>Content Path:</strong> The file path of a file on disk or in an object store, relative to the
<a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object-root">OCFL Object Root</a>. Content paths are used in the <a href="https://ocfl.io/1.1/spec/#dfn-manifest">Manifest</a> within an
<a href="https://ocfl.io/1.1/spec/#dfn-inventory">Inventory</a>.</p>
  </li>
  <li>
    <p><a name="dfn-digest"></a><strong>Digest:</strong> An algorithmic characterization of the contents of a file conforming to a standard
digest algorithm.</p>
  </li>
  <li>
    <p><a name="dfn-extension"></a><strong>Extension:</strong> Extensions are used to collaborate, review, and publish additional
non-normative functions related to OCFL. Extensions are intended to be informational and cite-able, but outside the
scope of the normal specification process. Registered extensions may be found in the <a href="https://ocfl.github.io/extensions/">OCFL Extensions
repository.</a></p>
  </li>
  <li>
    <p><a name="dfn-inventory"></a><strong>Inventory:</strong> A file, expressed in JSON, that tracks the history and current state of an
OCFL Object.</p>
  </li>
  <li>
    <p><a name="dfn-logical-path"></a><strong>Logical Path:</strong> A path that represents a file’s location in the <a href="https://ocfl.io/1.1/spec/#dfn-logical-state">logical
state</a> of an object. Logical paths are used in conjunction with a digest to represent the file name
and path for a given bitstream at a given version.</p>
  </li>
  <li>
    <p><a name="dfn-logical-state"></a><strong>Logical State:</strong> A grouping of logical paths tied to their corresponding bitstreams
that reflect the state of the object content for a given version.</p>
  </li>
  <li>
    <p><a name="dfn-logs-directory"></a><strong>Logs Directory:</strong> A directory for storing information about the content (e.g., actions
performed) that is not part of the content itself.</p>
  </li>
  <li>
    <p><a name="dfn-manifest"></a><strong>Manifest:</strong> A section of the <a href="https://ocfl.io/1.1/spec/#dfn-inventory">Inventory</a> listing all files and their digests
within an OCFL Object.</p>
  </li>
  <li>
    <p><a name="dfn-ocfl-object"></a><strong>OCFL Object:</strong> A group of one or more content files and administrative information, that
together have a unique identifier. The object may contain a sequence of versions of the files that represent the
evolution of the object’s contents.</p>
  </li>
  <li>
    <p><a name="dfn-ocfl-object-root"></a><strong>OCFL Object Root:</strong> The base directory of an <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object">OCFL Object</a>,
identified by a [<a href="https://ocfl.io/1.1/spec/#ref-namaste">NAMASTE</a>] file “0=ocfl_object_1.1”.</p>
  </li>
  <li>
    <p><a name="dfn-ocfl-storage-root"></a><strong>OCFL Storage Root:</strong> A base directory used to store OCFL Objects, identified by a
[<a href="https://ocfl.io/1.1/spec/#ref-namaste">NAMASTE</a>] file “0=ocfl_1.1”.</p>
  </li>
  <li>
    <p><a name="dfn-ocfl-version"></a><strong>OCFL Version:</strong> The state of an <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object">OCFL Object</a>’s content which is
constructed using the incremental changes recorded in the sequence of corresponding and prior version directories.</p>
  </li>
  <li>
    <p><a name="dfn-registered-extension-name"></a><strong>Registered Extension Name:</strong> The registered name of an extension is the
name provided in the <em>Extension Name</em> property of the extension’s definition in the <a href="https://ocfl.github.io/extensions/">OCFL Extensions
repository</a>.</p>
  </li>
</ul>

<h2 id="object-spec">3. OCFL Object<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#object-spec" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<p>An OCFL Object is a group of one or more content files and administrative information, that are together identified by a
URI. The object may contain a sequence of versions of the files that represent the evolution of the object’s contents.</p>

<p>A file is defined as a content bitstream that can be stored and transmitted. Directories (also called “folders”) allow
for the organization of files into tree-like hierarchies. The content of an OCFL Object is the files and the directories
they are organized in that are stored <em>within</em> the hierarchy layout described in this specification.</p>

<p>An OCFL Object includes administrative information that identifies a directory as an OCFL Object, and also provides a
means of tracking changes to the contents of the object over time.</p>

<p>An OCFL Object is therefore:</p>

<ol>
  <li>
    <p>A conceptual gathering of all files (data and metadata), the directories they are organized in, and their changes
over time which together form the digital representation of an entity that need to be managed, in preservation terms, as
a single coherent whole (i.e., content); and</p>
  </li>
  <li>
    <p>A file and directory layout and administrative information on a storage medium that provides a defined structure for
the storage of this content, and through which these files and their changes may be understood (i.e., structure).</p>
  </li>
</ol>

<p>A key goal of the OCFL is the rebuildability of a repository from an OCFL Storage Root without additional information
resources. Consequently, a key implementation consideration should be to ensure that OCFL Objects contain all the data
and metadata required to achieve this. With reference to the [<a href="https://ocfl.io/1.1/spec/#ref-oais">OAIS</a>] model, this would include all the
descriptive, administrative, structural, representation and preservation metadata relevant to the object.</p>

<p>A central feature of the OCFL specification is support for versioning. This recognizes that digital objects will change
over time, through new requirements, fixes, updates, or format shifts. The specification takes no position on what
constitutes a version or a versionable action, but it is recommended that implementers have a clear position on this
within their local storage policies.</p>

<h3 id="object-structure">3.1 Object Structure<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#object-structure" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The OCFL Object structure organizes content files and administrative information in order to support content storage and
object validation. The structure for an object with one version is shown in the following figure:</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[object_root]
    ├── 0=ocfl_object_1.1
    ├── inventory.json
    ├── inventory.json.sha512
    └── v1
        ├── inventory.json
        ├── inventory.json.sha512
        └── content
               └── ... content files ...
</code></pre></div></div>

<p>The <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object-root">OCFL Object Root</a> <span id="E001" class="rfc2119">MUST NOT</span> contain files or
directories other than those specified in the following sections (3.2 through 3.9).</p>

<h3 id="object-conformance-declaration">3.2 Object Conformance Declaration<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#object-conformance-declaration" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The OCFL specification version declaration <span id="E002" class="rfc2119">MUST</span> be formatted according to the
[<a href="https://ocfl.io/1.1/spec/#ref-namaste">NAMASTE</a>] specification. There <span id="E003" class="rfc2119">MUST</span> be exactly one version
declaration file in the base directory of the <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object-root">OCFL Object Root</a> giving the OCFL version in the
filename. The filename <span id="E004" class="rfc2119">MUST</span> conform to the pattern <code class="language-plaintext highlighter-rouge">T=dvalue</code>, where <code class="language-plaintext highlighter-rouge">T</code> <span id="E005" class="rfc2119">MUST</span> be 0, and <code class="language-plaintext highlighter-rouge">dvalue</code> <span id="E006" class="rfc2119">MUST</span> be <code class="language-plaintext highlighter-rouge">ocfl_object_</code>,
followed by the OCFL specification version number. The text contents of the file <span id="E007" class="rfc2119">MUST</span> be the same as <code class="language-plaintext highlighter-rouge">dvalue</code>, followed by a newline (<code class="language-plaintext highlighter-rouge">\n</code>).</p>

<h3 id="version-directories">3.3 Version Directories<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#version-directories" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>OCFL Object content <span id="E008" class="rfc2119">MUST</span> be stored as a sequence of one or more versions. Each
object version is stored in a version directory under the object root. Version directory names <span id="E104" class="rfc2119">MUST</span> be constructed by prepending <code class="language-plaintext highlighter-rouge">v</code> to the version number. The version number <span id="E105" class="rfc2119">MUST</span> be taken from the sequence of positive, base-ten integers: 1, 2, 3, etc.. The version number
sequence <span id="E009" class="rfc2119">MUST</span> start at 1 and <span id="E010" class="rfc2119">MUST</span> be
continuous without missing integers.</p>

<p>Implementations <span id="W001" class="rfc2119">SHOULD</span> use version directory names constructed without
zero-padding the version number, ie. <code class="language-plaintext highlighter-rouge">v1</code>, <code class="language-plaintext highlighter-rouge">v2</code>, <code class="language-plaintext highlighter-rouge">v3</code>, etc..</p>

<p>For compatibility with existing filesystem conventions, implementations <span class="rfc2119">MAY</span> use zero-padded
version directory numbers, with the following restriction: If zero-padded version directory numbers are used then they
<span id="E011" class="rfc2119">MUST</span> start with the prefix <code class="language-plaintext highlighter-rouge">v</code> and then a zero. For example, in an implementation
that uses five digits for version directory names then <code class="language-plaintext highlighter-rouge">v00001</code> to <code class="language-plaintext highlighter-rouge">v09999</code> are allowed, <code class="language-plaintext highlighter-rouge">v10000</code> is not allowed.</p>

<p>The first version of an object defines the naming convention for all version directories for the object. All version
directories of an object <span id="E012" class="rfc2119">MUST</span> use the same naming convention: either a non-padded
version directory number, or a zero-padded version directory number of consistent length. The version naming convention
<span id="E013" class="rfc2119">MUST</span> be consistent across all versions. In all cases, references to files inside
version directories from inventory files <span id="E014" class="rfc2119">MUST</span> use the actual version directory
names.</p>

<p>There <span id="E015" class="rfc2119">MUST</span> be no other files as children of a version directory, other than an
<a href="https://ocfl.io/1.1/spec/#inventory">inventory file</a> and a <a href="https://ocfl.io/1.1/spec/#inventory-digest">inventory digest</a>. The version directory <span id="W002" class="rfc2119">SHOULD NOT</span> contain any directories other than the designated content sub-directory. Once created,
the contents of a version directory are expected to be immutable.</p>

<h4 id="content-directory">3.3.1 Content Directory<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#content-directory" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>Version directories <span id="E016" class="rfc2119">MUST</span> contain a designated content sub-directory if the
version contains files to be preserved, and <span id="W003" class="rfc2119">SHOULD NOT</span> contain this sub-directory
otherwise. The name of this designated sub-directory <span class="rfc2119">MAY</span> be defined in the <a href="https://ocfl.io/1.1/spec/#inventory">inventory
file</a> using the key <code class="language-plaintext highlighter-rouge">contentDirectory</code> with the value being the chosen sub-directory name as a string,
relative to the version directory. The <code class="language-plaintext highlighter-rouge">contentDirectory</code> value <span id="E108" class="rfc2119">MUST</span> represent a
direct child directory of the version directory in which it is found. As such, the <code class="language-plaintext highlighter-rouge">contentDirectory</code> value <span id="E017" class="rfc2119">MUST NOT</span> contain the forward slash (<code class="language-plaintext highlighter-rouge">/</code>) path separator and <span id="E018" class="rfc2119">MUST NOT</span> be either one or two periods (<code class="language-plaintext highlighter-rouge">.</code> or <code class="language-plaintext highlighter-rouge">..</code>). If the key <code class="language-plaintext highlighter-rouge">contentDirectory</code> is set, it
<span id="E019" class="rfc2119">MUST</span> be set in the first version of the object and <span id="E020" class="rfc2119">MUST NOT</span> change between versions of the same object.</p>

<p>If the key <code class="language-plaintext highlighter-rouge">contentDirectory</code> is not present in the <a href="https://ocfl.io/1.1/spec/#inventory">inventory file</a> then the name of the designated content
sub-directory <span id="E021" class="rfc2119">MUST</span> be <code class="language-plaintext highlighter-rouge">content</code>. OCFL-compliant tools (including any validators)
<span id="E022" class="rfc2119">MUST</span> ignore all directories in the object version directory except for the
designated content directory.</p>

<p>Every file within a version’s content directory <span id="E023" class="rfc2119">MUST</span> be referenced in the
<a href="https://ocfl.io/1.1/spec/#manifest">manifest</a> section of that version’s inventory. There <span id="E024" class="rfc2119">MUST NOT</span> be
empty directories within a version’s content directory. A directory that would otherwise be empty <span class="rfc2119">MAY</span> be maintained by creating a file within it named according to local conventions, for example
by making an empty <code class="language-plaintext highlighter-rouge">.keep</code> file.</p>

<h3 id="digests">3.4 Digests<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#digests" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>A <a href="https://ocfl.io/1.1/spec/#dfn-digest">digest</a> plays two roles in an OCFL Object. The first is that digests allow for content-addressable
reference to files within the OCFL Object. That is, the connection between a file’s <a href="https://ocfl.io/1.1/spec/#dfn-content-path">content path</a> on
physical storage and its <a href="https://ocfl.io/1.1/spec/#dfn-logical-path">logical path</a> in a version of the object’s content is made with a digest of
its contents, rather than its filename. This use of the content digest facilitates de-duplication of files with the same
content within an object, such as files that are unchanged from one version to the next. The second role that digests
play is provide for fixity checks to determine whether a file has become corrupt, through hardware degradation or
accident for example.</p>

<p>For content-addressing, OCFL Objects <span id="E025" class="rfc2119">MUST</span> use either <code class="language-plaintext highlighter-rouge">sha512</code> or <code class="language-plaintext highlighter-rouge">sha256</code>, and
<span id="W004" class="rfc2119">SHOULD</span> use <code class="language-plaintext highlighter-rouge">sha512</code>. The choice of the <code class="language-plaintext highlighter-rouge">sha512</code> digest algorithm as default
recognizes that it has no known collision vulnerabilities and multiple implementations are available.</p>

<p>For storage of additional fixity values, or to support legacy content migration, implementers <span id="E026" class="rfc2119">MUST</span> choose from the following controlled vocabulary of digest algorithms, or from a list of
additional algorithms given in the [<a href="https://ocfl.io/1.1/spec/#ref-digest-algorithms-extension">Digest-Algorithms-Extension</a>]. OCFL clients
<span id="E027" class="rfc2119">MUST</span> support all fixity algorithms given in the table below, and <span class="rfc2119">MAY</span> support additional algorithms from the extensions. Optional fixity algorithms that are not
supported by a client <span id="E028" class="rfc2119">MUST</span> be ignored by that client.</p>

<table>
  <thead>
    <tr>
      <th>Digest Algorithm Name</th>
      <th>Note</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code class="language-plaintext highlighter-rouge">md5</code></td>
      <td>Insecure. Use only for legacy fixity values. MD5 algorithm and hex encoding defined by [<a href="https://ocfl.io/1.1/spec/#ref-rfc1321">RFC1321</a>]. For example, the <code class="language-plaintext highlighter-rouge">md5</code> digest of a zero-length bitstream is <code class="language-plaintext highlighter-rouge">d41d8cd98f00b204e9800998ecf8427e</code>.</td>
    </tr>
    <tr>
      <td><code class="language-plaintext highlighter-rouge">sha1</code></td>
      <td>Insecure. Use only for legacy fixity values. SHA-1 algorithm defined by [<a href="https://ocfl.io/1.1/spec/#ref-fips-180-4">FIPS-180-4</a>] and <span id="E029" class="rfc2119">MUST</span> be encoded using hex (base16) encoding [<a href="https://ocfl.io/1.1/spec/#ref-rfc4648">RFC4648</a>]. For example, the <code class="language-plaintext highlighter-rouge">sha1</code> digest of a zero-length bitstream is <code class="language-plaintext highlighter-rouge">da39a3ee5e6b4b0d3255bfef95601890afd80709</code>.</td>
    </tr>
    <tr>
      <td><code class="language-plaintext highlighter-rouge">sha256</code></td>
      <td>Non-truncated form only; note performance implications. SHA-256 algorithm defined by [<a href="https://ocfl.io/1.1/spec/#ref-fips-180-4">FIPS-180-4</a>] and <span id="E030" class="rfc2119">MUST</span> be encoded using hex (base16) encoding [<a href="https://ocfl.io/1.1/spec/#ref-rfc4648">RFC4648</a>]. For example, the <code class="language-plaintext highlighter-rouge">sha256</code> digest of a zero-length bitstream starts <code class="language-plaintext highlighter-rouge">e3b0c44298fc1c149afbf4c8996fb92427ae41e4...</code> (64 hex digits long).</td>
    </tr>
    <tr>
      <td><code class="language-plaintext highlighter-rouge">sha512</code></td>
      <td>Default choice. Non-truncated form only. SHA-512 algorithm defined by [<a href="https://ocfl.io/1.1/spec/#ref-fips-180-4">FIPS-180-4</a>] and <span id="E031" class="rfc2119">MUST</span> be encoded using hex (base16) encoding [<a href="https://ocfl.io/1.1/spec/#ref-rfc4648">RFC4648</a>]. For example, the <code class="language-plaintext highlighter-rouge">sha512</code> digest of a zero-length bitstream starts <code class="language-plaintext highlighter-rouge">cf83e1357eefb8bdf1542850d66d8007d620e405...</code> (128 hex digits long).</td>
    </tr>
    <tr>
      <td><code class="language-plaintext highlighter-rouge">blake2b-512</code></td>
      <td>Full-length form only, using the 2B variant (64 bit) as defined by [<a href="https://ocfl.io/1.1/spec/#ref-rfc7693">RFC7693</a>]. <span id="E032" class="rfc2119">MUST</span> be encoded using hex (base16) encoding [<a href="https://ocfl.io/1.1/spec/#ref-rfc4648">RFC4648</a>]. For example, the <code class="language-plaintext highlighter-rouge">blake2b-512</code> digest of a zero-length bitstream starts <code class="language-plaintext highlighter-rouge">786a02f742015903c6c6fd852552d272912f4740...</code> (128 hex digits long).</td>
    </tr>
  </tbody>
</table>

<p>An OCFL Inventory <span class="rfc2119">MAY</span> contain a fixity section that can store one or more blocks containing
fixity values using multiple digest algorithms. See the <a href="https://ocfl.io/1.1/spec/#fixity">section on fixity</a> below for further details.</p>

<blockquote>
  <p>Non-normative note: Implementers may also store copies of their file digests in a system external to their OCFL Object
stores at the point of ingest, to further safeguard against the possibility of malicious manipulation of file contents
and digests.</p>

  <p>Implementers should be aware that base16 digests are case insensitive. Different tools will generate digests in
uppercase or lowercase, and this may lead to case differences between references to a digest and the digest itself
within the inventory. If string-based methods are used to work with digests and inventories (as is the case in most
common JSON libraries) then extra care must be taken to ensure case-insensitive comparisons are being made.</p>
</blockquote>

<h3 id="inventory">3.5 Inventory<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#inventory" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>An OCFL Object Inventory <span id="E033" class="rfc2119">MUST</span> follow the JSON (defined by
[<a href="https://ocfl.io/1.1/spec/#ref-rfc8259">RFC8259</a>]) structure described in this section with contents encoded in UTF-8, and <span id="E034" class="rfc2119">MUST</span> be named <code class="language-plaintext highlighter-rouge">inventory.json</code>. The order of entries in both the JSON objects and arrays used in
inventory files has no significance. An OCFL Object Inventory <span id="E102" class="rfc2119">MUST NOT</span> contain
any keys not described in this specification.</p>

<p>The forward slash (/) path separator <span id="E035" class="rfc2119">MUST</span> be used in content paths in the
<a href="https://ocfl.io/1.1/spec/#manifest">manifest</a> and <a href="https://ocfl.io/1.1/spec/#fixity">fixity</a> blocks within the inventory. Implementations that target systems using other
separators will need to translate paths appropriately.</p>

<blockquote>
  <p>Non-normative note: A [<a href="https://ocfl.io/1.1/spec/#ref-json-schema">JSON-Schema</a>] for validating OCFL Object Inventory files is provided at
<a href="https://ocfl.io/1.1/spec/inventory_schema.json">inventory_schema.json</a>.</p>
</blockquote>

<h4 id="inventory-structure">3.5.1 Basic Structure<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#inventory-structure" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>Every OCFL inventory <span id="E036" class="rfc2119">MUST</span> include the following keys:</p>

<ul>
  <li>
    <p><code class="language-plaintext highlighter-rouge">id</code>: A unique identifier for the OCFL Object. This <span id="E037" class="rfc2119">MUST</span> be unique in the local
context, <span id="E110" class="rfc2119">MUST NOT</span> change between versions of the same object, and <span id="W005" class="rfc2119">SHOULD</span> be a URI [<a href="https://ocfl.io/1.1/spec/#ref-rfc3986">RFC3986</a>]. There is no expectation that a URI used is
resolvable. For example, URNs [<a href="https://ocfl.io/1.1/spec/#ref-rfc8141">RFC8141</a>] <span class="rfc2119">MAY</span> be used.</p>
  </li>
  <li>
    <p><code class="language-plaintext highlighter-rouge">type</code>: A type for the inventory JSON object that also serves to document the OCFL specification version that the
inventory complies with. In the object root inventory this <span id="E038" class="rfc2119">MUST</span> be the URI of the
inventory section of the specification version matching the object conformance declaration. For the current
specification version the value is <code class="language-plaintext highlighter-rouge">https://ocfl.io/1.1/spec/#inventory</code>.</p>
  </li>
  <li>
    <p><code class="language-plaintext highlighter-rouge">digestAlgorithm</code>: The digest algorithm used for calculating digests for content-addressing within the OCFL Object and
for the <a href="https://ocfl.io/1.1/spec/#inventory-digest">Inventory Digest</a>. This <span id="E039" class="rfc2119">MUST</span> be the algorithm used in
the <code class="language-plaintext highlighter-rouge">manifest</code> and <code class="language-plaintext highlighter-rouge">state</code> blocks, see the <a href="https://ocfl.io/1.1/spec/#digests">section on Digests</a> for more information about algorithms.</p>
  </li>
  <li>
    <p><code class="language-plaintext highlighter-rouge">head</code>: The version directory name of the most recent version of the object. This <span id="E040" class="rfc2119">MUST</span> be the version directory name with the highest version number.</p>
  </li>
</ul>

<p>There <span class="rfc2119">MAY</span> be the following key:</p>

<ul>
  <li><code class="language-plaintext highlighter-rouge">contentDirectory</code>: The name of the designated content directory within the version directories. If not specified then
the content directory name is <code class="language-plaintext highlighter-rouge">content</code>.</li>
</ul>

<p>In addition to these keys, there <span id="E041" class="rfc2119">MUST</span> be two other blocks present, <code class="language-plaintext highlighter-rouge">manifest</code> and
<code class="language-plaintext highlighter-rouge">versions</code>, which are discussed in the next two sections.</p>

<h4 id="manifest">3.5.2 Manifest<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#manifest" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>The value of the <code class="language-plaintext highlighter-rouge">manifest</code> key <span id="E106" class="rfc2119">MUST</span> be a JSON object, and each key <span id="E107" class="rfc2119">MUST</span> correspond to a digest value key found in one or more <code class="language-plaintext highlighter-rouge">state</code> blocks of the
current and/or previous <code class="language-plaintext highlighter-rouge">version</code> blocks of the <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object">OCFL Object</a>. The value for each key <span id="E092" class="rfc2119">MUST</span> be an array containing the <a href="https://ocfl.io/1.1/spec/#dfn-content-path">content path</a>s of files in the OCFL Object
that have content with the given digest. As JSON keys are case sensitive, for digest algorithms with case insensitive
digest values, there is an additional requirement that each digest value <span id="E096" class="rfc2119">MUST</span>
occur only once in the manifest block for any digest algorithm, regardless of case. Content paths within a manifest
block <span id="E042" class="rfc2119">MUST</span> be relative to the <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object-root">OCFL Object Root</a>. The
following restrictions avoid ambiguity and provide path safety for clients processing the <code class="language-plaintext highlighter-rouge">manifest</code>.</p>

<ul>
  <li>
    <p>The content path <span id="E098" class="rfc2119">MUST</span> be interpreted as a set of one or more path elements
joined by a <code class="language-plaintext highlighter-rouge">/</code> path separator.</p>
  </li>
  <li>
    <p>Path elements <span id="E099" class="rfc2119">MUST NOT</span> be <code class="language-plaintext highlighter-rouge">.</code>, <code class="language-plaintext highlighter-rouge">..</code>, or empty (<code class="language-plaintext highlighter-rouge">//</code>).</p>
  </li>
  <li>
    <p>A content path <span id="E100" class="rfc2119">MUST NOT</span> begin or end with a forward slash (<code class="language-plaintext highlighter-rouge">/</code>).</p>
  </li>
  <li>
    <p>Within an inventory, content paths <span id="E101" class="rfc2119">MUST</span> be unique and non-conflicting, so the
content path for a file cannot appear as the initial part of another content path.</p>
  </li>
</ul>

<blockquote>
  <p>Non-normative note: If only one file is stored in the OCFL Object for each digest, fully de-duplicating the content,
then there will be only one <a href="https://ocfl.io/1.1/spec/#dfn-content-path">content path</a> for each digest. There may, however, be multiple logical
paths for a given digest if the content was not entirely de-duplicated when constructing the OCFL Object.</p>

  <p>An example manifest object for three content paths, all in version 1, is shown below:</p>

  <div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="nl">"manifest"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
  </span><span class="nl">"7dcc35...c31"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
  </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/empty.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
  </span><span class="nl">"ffccf6...62e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/image.tiff"</span><span class="w"> </span><span class="p">]</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div>  </div>
</blockquote>

<h4 id="versions">3.5.3 Versions<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#versions" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>An OCFL Object Inventory <span id="E043" class="rfc2119">MUST</span> include a block for storing versions. This block
<span id="E044" class="rfc2119">MUST</span> have the key of <code class="language-plaintext highlighter-rouge">versions</code> within the inventory, and it <span id="E045" class="rfc2119">MUST</span> be a JSON object. The keys of this object <span id="E046" class="rfc2119">MUST</span>
correspond to the names of the <a href="https://ocfl.io/1.1/spec/#version-directories">version directories</a> used. Each value <span id="E047" class="rfc2119">MUST</span> be another JSON object that characterizes the version, as described in the <a href="https://ocfl.io/1.1/spec/#version">3.5.3.1
Version</a> section.</p>

<h5 id="version">3.5.3.1 Version<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#version" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h5>

<p>A JSON object to describe one <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-version">OCFL Version</a>, which <span id="E048" class="rfc2119">MUST</span>
include the following keys:</p>

<ul>
  <li>
    <p><code class="language-plaintext highlighter-rouge">created</code>: The value of this key is the datetime of creation of this version. It <span id="E049" class="rfc2119">MUST</span> be expressed in the Internet Date/Time Format defined by [<a href="https://ocfl.io/1.1/spec/#ref-rfc3339">RFC3339</a>]. This
format requires the inclusion of a timezone value or <code class="language-plaintext highlighter-rouge">Z</code> for UTC, and that the time component be granular to the second
level (with optional fractional seconds).</p>
  </li>
  <li>
    <p><code class="language-plaintext highlighter-rouge">state</code>: The value of this key is a JSON object, containing a list of keys and values corresponding to the <a href="https://ocfl.io/1.1/spec/#dfn-logical-state">logical
state</a> of the object at that version. The keys of this JSON object are digest values, each of which
<span id="E050" class="rfc2119">MUST</span> exactly match a digest value key in the <a href="https://ocfl.io/1.1/spec/#manifest">manifest of the
inventory</a>. The value for each key is an array containing <a href="https://ocfl.io/1.1/spec/#dfn-logical-path">logical path</a> names of files in
the OCFL Object’s logical state that have content with the given digest.</p>
  </li>
</ul>

<p><a href="https://ocfl.io/1.1/spec/#logical-path">Logical paths</a> present the structure of an OCFL Object at a given version. This is given as an array of
values, with the following restrictions to provide for path safety in the common case of the logical path value
representing a file path.</p>

<ul>
  <li>
    <p>The logical path <span id="E051" class="rfc2119">MUST</span> be interpreted as a set of one or more path elements
joined by a <code class="language-plaintext highlighter-rouge">/</code> path separator.</p>
  </li>
  <li>
    <p>Path elements <span id="E052" class="rfc2119">MUST NOT</span> be <code class="language-plaintext highlighter-rouge">.</code>, <code class="language-plaintext highlighter-rouge">..</code>, or empty (<code class="language-plaintext highlighter-rouge">//</code>).</p>
  </li>
  <li>
    <p>A logical path <span id="E053" class="rfc2119">MUST NOT</span> begin or end with a forward slash (<code class="language-plaintext highlighter-rouge">/</code>).</p>
  </li>
  <li>
    <p>Within a version, logical paths <span id="E095" class="rfc2119">MUST</span> be unique and non-conflicting, so the
logical path for a file cannot appear as the initial part of another logical path.</p>
  </li>
</ul>

<blockquote>
  <p>Non-normative note: The <a href="https://ocfl.io/1.1/spec/#dfn-logical-state">logical state</a> of the object uses content-addressing to map logical paths
to their bitstreams, as expressed in the manifest section of the inventory. Notably, the version state provides
de-duplication of content within the OCFL Object by mapping multiple logical paths with the same content to the same
digest in the manifest. See [<a href="https://ocfl.io/1.1/spec/#ref-ocfl-implementation-notes">OCFL-Implementation-Notes</a>].</p>

  <p>An example <code class="language-plaintext highlighter-rouge">state</code> block is shown below:</p>

  <div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
  </span><span class="nl">"4d27c8...b53"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
  </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"empty.txt"</span><span class="p">,</span><span class="w"> </span><span class="s2">"empty2.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div>  </div>

  <p>This <code class="language-plaintext highlighter-rouge">state</code> block describes an object with 3 files, two of which have the same content (<code class="language-plaintext highlighter-rouge">empty.txt</code> and
<code class="language-plaintext highlighter-rouge">empty2.txt</code>), and one of which is in a sub-directory (<code class="language-plaintext highlighter-rouge">bar.xml</code>). The <a href="https://ocfl.io/1.1/spec/#dfn-logical-state">logical state</a> shown as a
tree is thus:</p>

  <div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>├── empty.txt
├── empty2.txt
└── foo
    └── bar.xml
</code></pre></div>  </div>
</blockquote>

<p>The JSON object describing an <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-version">OCFL Version</a>, <span id="W007" class="rfc2119">SHOULD</span> include
the following keys:</p>

<ul>
  <li>
    <p><code class="language-plaintext highlighter-rouge">message</code>: The value of this key is freeform text, used to record the rationale for creating this version. It <span id="E094" class="rfc2119">MUST</span> be a JSON string.</p>
  </li>
  <li>
    <p><code class="language-plaintext highlighter-rouge">user</code>: The value of this key is a JSON object intended to identify the user or agent that created the current <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-version">OCFL
Version</a>. The value of the <code class="language-plaintext highlighter-rouge">user</code> key <span id="E054" class="rfc2119">MUST</span> contain a user name
key, <code class="language-plaintext highlighter-rouge">name</code> and <span id="W008" class="rfc2119">SHOULD</span> contain an address key, <code class="language-plaintext highlighter-rouge">address</code>. The <code class="language-plaintext highlighter-rouge">name</code> value is any
readable name of the user, e.g., a proper name, user ID, agent ID. The <code class="language-plaintext highlighter-rouge">address</code> value <span id="W009" class="rfc2119">SHOULD</span> be a URI: either a mailto URI [<a href="https://ocfl.io/1.1/spec/#ref-rfc6068">RFC6068</a>] with the e-mail address of the
user or a URL to a personal identifier, e.g., an ORCID iD.</p>
  </li>
</ul>

<h4 id="fixity">3.5.4 Fixity<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#fixity" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>An OCFL Object inventory <span class="rfc2119">MAY</span> include a block for storing additional fixity information to
supplement the complete set of digests in the <a href="https://ocfl.io/1.1/spec/#manifest">Manifest</a>, for example to support legacy digests from a
content migration. If present, this block <span id="E055" class="rfc2119">MUST</span> have the key of <code class="language-plaintext highlighter-rouge">fixity</code> within
the inventory, and its value <span id="E111" class="rfc2119">MUST</span> be a JSON object, which <span class="rfc2119">MAY</span> be empty.</p>

<p>The keys within the <code class="language-plaintext highlighter-rouge">fixity</code> block <span id="E056" class="rfc2119">MUST</span> correspond to the controlled vocabulary
of <a href="https://ocfl.io/1.1/spec/#digest-algorithms">digest algorithm names</a> listed in the <a href="https://ocfl.io/1.1/spec/#digests">Digests</a> section, or in a table given in an
<a href="https://ocfl.io/1.1/spec/#dfn-extension">Extension</a>. The value of the fixity block for a particular digest algorithm <span id="E057" class="rfc2119">MUST</span> follow the structure of the <a href="https://ocfl.io/1.1/spec/#manifest">3.5.2 Manifest</a> block; that is, a key corresponding
to the digest value, and an array of <a href="https://ocfl.io/1.1/spec/#dfn-content-path">content path</a>s. The <code class="language-plaintext highlighter-rouge">fixity</code> block for any digest algorithm
<span class="rfc2119">MAY</span> include digest values for any subset of content paths in the object. Where included,
the digest values given <span id="E093" class="rfc2119">MUST</span> match the digests of the files at the corresponding
content paths. As JSON keys are case sensitive, for digest algorithms with case insensitive digest values, there is an
additional requirement that each digest value <span id="E097" class="rfc2119">MUST</span> occur only once in the
<code class="language-plaintext highlighter-rouge">fixity</code> block for any digest algorithm, regardless of case. There is no requirement that all content files have a value
in the <code class="language-plaintext highlighter-rouge">fixity</code> block, or that fixity values provided in one version are carried forward to later versions.</p>

<blockquote>
  <p>An example <code class="language-plaintext highlighter-rouge">fixity</code> with <code class="language-plaintext highlighter-rouge">md5</code> and <code class="language-plaintext highlighter-rouge">sha1</code> digests is shown below. In this case the <code class="language-plaintext highlighter-rouge">md5</code> digest values are provided
only for version 1 content paths.</p>

  <div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="nl">"fixity"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
  </span><span class="nl">"md5"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"184f84e28cbe75e050e9c25ea7f2e939"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"c289c8ccd4bab6e385f5afdd89b5bda2"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/image.tiff"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"d41d8cd98f00b204e9800998ecf8427e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/empty.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
  </span><span class="p">},</span><span class="w">
  </span><span class="nl">"sha1"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"66709b068a2faead97113559db78ccd44712cbf2"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"a6357c99ecc5752931e133227581e914968f3b9c"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v2/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"b9c7ccc6154974288132b63c15db8d2750716b49"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/image.tiff"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"da39a3ee5e6b4b0d3255bfef95601890afd80709"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/empty.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
  </span><span class="p">}</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div>  </div>
</blockquote>

<h3 id="inventory-digest">3.6 Inventory Digest<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#inventory-digest" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>Every occurrence of an inventory file <span id="E058" class="rfc2119">MUST</span> have an accompanying sidecar file
named <code class="language-plaintext highlighter-rouge">inventory.json.ALGORITHM</code> stating its digest, where <code class="language-plaintext highlighter-rouge">ALGORITHM</code> is the chosen digest algorithm for the object.
The ALGORITHM <span id="E059" class="rfc2119">MUST</span> match the value given for the <code class="language-plaintext highlighter-rouge">digestAlgorithm</code> key in the
inventory. An example might be <code class="language-plaintext highlighter-rouge">inventory.json.sha512</code>.</p>

<p>The digest sidecar file <span id="E060" class="rfc2119">MUST</span> contain the digest of the inventory file. This <span id="E061" class="rfc2119">MUST</span> follow the format:</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>DIGEST inventory.json
</code></pre></div></div>

<p>One or more whitespace characters (spaces or tabs) must separate DIGEST from the string <code class="language-plaintext highlighter-rouge">inventory.json</code>; that is, the
name of the inventory file in the same directory.</p>

<p>The digest of the inventory <span id="E062" class="rfc2119">MUST</span> be computed only after all changes to the
inventory have been made, and thus writing the digest sidecar file is the last step in the versioning process.</p>

<h3 id="version-inventory">3.7 Version Inventory and Inventory Digest<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#version-inventory" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>Every OCFL Object <span id="E063" class="rfc2119">MUST</span> have an inventory file within the OCFL Object Root,
corresponding to the state of the OCFL Object at the current version. Additionally, every version directory <span id="W010" class="rfc2119">SHOULD</span> include an inventory file that is an <a href="https://ocfl.io/1.1/spec/#inventory">Inventory</a> of all content for
versions up to and including that particular version. Where an OCFL Object contains <code class="language-plaintext highlighter-rouge">inventory.json</code> in version
directories, the inventory file in the OCFL Object Root <span id="E064" class="rfc2119">MUST</span> be the same as the
file in the most recent version. See also requirements for the corresponding <a href="https://ocfl.io/1.1/spec/#inventory-digest">Inventory Digest</a>.</p>

<p>In the case that prior version directories include an inventory file there will be multiple inventory files describing
prior versions within the OCFL Object. Each <code class="language-plaintext highlighter-rouge">version</code> block in each prior inventory file <span id="E066" class="rfc2119">MUST</span> represent the same <a href="https://ocfl.io/1.1/spec/#dfn-logical-state">logical state</a> as the corresponding <code class="language-plaintext highlighter-rouge">version</code> block
in the current inventory file. Additionally, the values of the <code class="language-plaintext highlighter-rouge">created</code>, <code class="language-plaintext highlighter-rouge">message</code> and <code class="language-plaintext highlighter-rouge">user</code> keys in each <code class="language-plaintext highlighter-rouge">version</code>
block in each prior inventory file <span id="W011" class="rfc2119">SHOULD</span> have the same values as the
corresponding keys in the corresponding <code class="language-plaintext highlighter-rouge">version</code> block in the current inventory file.</p>

<blockquote>
  <p>Non-normative note: Storing an inventory for every version provides redundancy for this critical information in a way
that is compatible with storage strategies that have immutable version directories.</p>
</blockquote>

<h4 id="conformance-of-prior-versions">3.7.1 Conformance of prior versions<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#conformance-of-prior-versions" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h4>

<p>Version directories in OCFL are intended to be immutable in that existing version directories do not change when a new
version directory is added. Each version directory within an OCFL Object <span id="E103" class="rfc2119">MUST</span>
conform to either the same or a later OCFL specification version as the preceding version directory. If inventories are
stored in the version directories then the OCFL specification version for a given version directory is apparent from the
<code class="language-plaintext highlighter-rouge">type</code> attribute in that <a href="https://ocfl.io/1.1/spec/#inventory-structure">inventory</a>.</p>

<h3 id="logs-directory">3.8 Logs Directory<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#logs-directory" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The base directory of an OCFL Object <span class="rfc2119">MAY</span> contain a directory named <code class="language-plaintext highlighter-rouge">logs</code>, which <span class="rfc2119">MAY</span> be empty. Implementers <span id="W012" class="rfc2119">SHOULD</span> use the <a href="https://ocfl.io/1.1/spec/#dfn-logs-directory">logs
directory</a> for storing files that contain a record of actions taken on the object. Since these logs
may be subject to local standards requirements, the format of these logs is considered out-of-scope for the OCFL Object.
Clients operating on the object <span class="rfc2119">MAY</span> log actions here that are not otherwise captured.</p>

<blockquote>
  <p>Non-normative note: The purpose of the logs directory is to provide implementers with a location for storing local
information about actions to the OCFL Object’s content that is not part of the content itself.</p>

  <p>As an example, implementers may have different local requirements to store audit information for their content. Some
may wish to store a log entry indicating that an audit was conducted, and nothing was wrong, while others may wish to
only store a log entry if an intervention was required.</p>
</blockquote>

<h3 id="object-extensions">3.9 Object Extensions<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#object-extensions" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The base directory of an OCFL Object <span class="rfc2119">MAY</span> contain a directory named <code class="language-plaintext highlighter-rouge">extensions</code> for the
purposes of extending the functionality of an OCFL Object. The <code class="language-plaintext highlighter-rouge">extensions</code> directory <span id="E067" class="rfc2119">MUST NOT</span> contain any files or sub-directories other than extension sub-directories.
Extension sub-directories <span id="W013" class="rfc2119">SHOULD</span> be named according to a <a href="https://ocfl.io/1.1/spec/#dfn-registered-extension-name">registered extension
name</a> in the <a href="https://ocfl.github.io/extensions/">OCFL Extensions repository</a>.</p>

<blockquote>
  <p>Non-normative note: Extension sub-directories should use the same name as a registered extension in order to both
avoid the possiblity of an extension sub-directory colliding with the name of another registered extension as well as to
facilitate the recognition of extensions by OCFL clients. See also <a href="https://ocfl.io/1.1/spec/#documenting-local-extensions">Documenting Local
Extensions</a>.</p>
</blockquote>

<h2 id="storage-root">4. OCFL Storage Root<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#storage-root" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<p>An <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-storage-root">OCFL Storage Root</a> is the base directory of an OCFL storage layout.</p>

<h3 id="root-structure">4.1 Root Structure<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#root-structure" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>An OCFL Storage Root <span id="E069" class="rfc2119">MUST</span> contain a <a href="https://ocfl.io/1.1/spec/#root-conformance-declaration">Root Conformance
Declaration</a> identifying it as such.</p>

<p>An OCFL Storage Root <span class="rfc2119">MAY</span> contain other files as direct children. These might include a
human-readable copy of the OCFL specification to make the storage root self-documenting, or files used to <a href="https://ocfl.io/1.1/spec/#documenting-local-extensions">document
local extensions</a>. The source file for this specification document is in
Markdown (described in [<a href="https://ocfl.io/1.1/spec/#ref-rfc7764">RFC7764</a>], which is designed to be readable as plain text as well as for
rendering as HTML, and thus makes it suitable for self-documentation. An OCFL validator <span id="E087" class="rfc2119">MUST</span> ignore any files in the storage root it does not understand.</p>

<p>An OCFL Storage Root <span id="E088" class="rfc2119">MUST NOT</span> contain directories or sub-directories other than
as a directory hierarchy used to store OCFL Objects or for <a href="https://ocfl.io/1.1/spec/#storage-root-extensions">storage root extensions</a>. The
directory hierarchy used to store OCFL Objects <span id="E072" class="rfc2119">MUST NOT</span> contain files that are
not part of an OCFL Object. Empty directories <span id="E073" class="rfc2119">MUST NOT</span> appear under a storage
root.</p>

<p>An OCFL Storage Root <span class="rfc2119">MAY</span> contain a file named <code class="language-plaintext highlighter-rouge">ocfl_layout.json</code> to describe the
arrangement of directories and OCFL objects under the storage root. If present, <code class="language-plaintext highlighter-rouge">ocfl_layout.json</code> <span id="E070" class="rfc2119">MUST</span> be a JSON (defined by [<a href="https://ocfl.io/1.1/spec/#ref-rfc8259">RFC8259</a>]) document encoded in UTF-8 and include the
following two keys in the root JSON object:</p>

<ul>
  <li>
    <p><code class="language-plaintext highlighter-rouge">extension</code> - An extension name that identifies an arrangement of directories and OCFL objects under the storage root,
i.e. how OCFL object identifiers are mapped to directory hierarchies. The value of the <code class="language-plaintext highlighter-rouge">extension</code> key <span id="E071" class="rfc2119">MUST</span> be the <a href="https://ocfl.io/1.1/spec/#dfn-registered-extension-name">registered extension name</a> for the extension
defining the arrangement under the storage root.</p>
  </li>
  <li>
    <p><code class="language-plaintext highlighter-rouge">description</code> - A human readable description of the arrangement of directories and OCFL objects under the storage
root.</p>
  </li>
</ul>

<p>Although implementations may require multiple OCFL Storage Roots—that is, several logical or physical volumes, or
multiple “buckets” in an object store—each OCFL Storage Root <span id="E074" class="rfc2119">MUST</span> be
independent.</p>

<p>The following example OCFL Storage Root represents the minimal set of files and folders:</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[storage_root]
    ├── 0=ocfl_1.1
    ├── ocfl_1.1.md        (human-readable text of the OCFL specification; optional)
    └── ocfl_layout.json   (description of storage hierarchy layout; optional)
</code></pre></div></div>

<h3 id="root-conformance-declaration">4.2 Root Conformance Declaration<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#root-conformance-declaration" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The OCFL version declaration <span id="E075" class="rfc2119">MUST</span> be formatted according to the
[<a href="https://ocfl.io/1.1/spec/#ref-namaste">NAMASTE</a>] specification. There <span id="E076" class="rfc2119">MUST</span> be exactly one version
declaration file in the base directory of the <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-storage-root">OCFL Storage Root</a> giving the OCFL version in the
filename. The filename <span id="E077" class="rfc2119">MUST</span> conform to the pattern <code class="language-plaintext highlighter-rouge">T=dvalue</code>, where <code class="language-plaintext highlighter-rouge">T</code> <span id="E078" class="rfc2119">MUST</span> be 0, and <code class="language-plaintext highlighter-rouge">dvalue</code> <span id="E079" class="rfc2119">MUST</span> be <code class="language-plaintext highlighter-rouge">ocfl_</code>,
followed by the OCFL specification version number. The text contents of the file <span id="E080" class="rfc2119">MUST</span> be the same as <code class="language-plaintext highlighter-rouge">dvalue</code>, followed by a newline (<code class="language-plaintext highlighter-rouge">\n</code>).</p>

<p>Root conformance indicates that the OCFL Storage Root conforms to this section (i.e. the OCFL Storage Root section) of
the specification. OCFL Objects within the OCFL Storage Root also include a conformance declaration which <span id="E081" class="rfc2119">MUST</span> indicate OCFL Object conformance to the same or earlier version of the
specification.</p>

<h3 id="root-hierarchies">4.3 Storage Hierarchies<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#root-hierarchies" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p><a href="https://ocfl.io/1.1/spec/#dfn-ocfl-object-root">OCFL Object Root</a>s <span id="E082" class="rfc2119">MUST</span> be stored either as the terminal
resource at the end of a directory storage hierarchy or as direct children of a containing <a href="https://ocfl.io/1.1/spec/#dfn-ocfl-storage-root">OCFL Storage
Root</a>.</p>

<p>A common practice is to use a unique identifier scheme to compose this storage hierarchy, typically arranged according
to some form of the [<a href="https://ocfl.io/1.1/spec/#ref-pairtree">PairTree</a>] specification. Irrespective of the pattern chosen for the storage
hierarchies, the following restrictions apply:</p>

<ol>
  <li>
    <p>There <span id="E083" class="rfc2119">MUST</span> be a deterministic mapping from an object identifier to a unique
storage path</p>
  </li>
  <li>
    <p>Storage hierarchies <span id="E084" class="rfc2119">MUST NOT</span> include files within intermediate directories</p>
  </li>
  <li>
    <p>Storage hierarchies <span id="E085" class="rfc2119">MUST</span> be terminated by OCFL Object Roots</p>
  </li>
  <li>
    <p>Storage hierarchies within the same OCFL Storage Root <span id="W014" class="rfc2119">SHOULD</span> use just one
layout pattern</p>
  </li>
  <li>
    <p>Storage hierarchies within the same OCFL Storage Root <span id="W015" class="rfc2119">SHOULD</span> consistently use
either a directory hierarchy of OCFL Objects or top-level OCFL Objects</p>
  </li>
</ol>

<h3 id="storage-root-extensions">4.4 Storage Root Extensions<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#storage-root-extensions" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The behavior of the storage root may be extended to support features from other specifications.</p>

<p>The base directory of an OCFL Storage Root <span class="rfc2119">MAY</span> contain a directory named <code class="language-plaintext highlighter-rouge">extensions</code> for
the purposes of extending the functionality of an OCFL Storage Root. The guidelines and limitations for the storage
root <code class="language-plaintext highlighter-rouge">extensions</code> directory are defined in alignment with those of the <a href="https://ocfl.io/1.1/spec/#object-extensions">object extensions</a>.</p>

<p>The <code class="language-plaintext highlighter-rouge">extensions</code> directory <span id="E112" class="rfc2119">MUST NOT</span> contain any files or sub-directories
other than extension sub-directories. Extension sub-directories <span id="W016" class="rfc2119">SHOULD</span> be named
according to a <a>registered extension name</a>.</p>

<blockquote>
  <p>Non-normative notes: Extension sub-directories should use the same name as a registered extension in order to both
avoid the possiblity of an extension sub-directory colliding with the name of another registered extension as well as to
facilitate the recognition of extensions by OCFL clients. See also <a href="https://ocfl.io/1.1/spec/#documenting-local-extensions">Documenting Local
Extensions</a>.</p>

  <p>Storage extensions can be used to support additional features, such as providing the storage
hierarchy disposition when pairtree is in use, or additional human-readable text about the nature of the storage root.</p>
</blockquote>

<h3 id="documenting-local-extensions">4.5 Documenting Local Extensions<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#documenting-local-extensions" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>It is preferable that both <a href="https://ocfl.io/1.1/spec/#object-extensions">Object Extensions</a> and <a href="https://ocfl.io/1.1/spec/#storage-root-extensions">Storage Root
Extenstions</a> are documented and registered in the <a href="https://ocfl.github.io/extensions/">OCFL Extensions
repository</a>. However, local extensions <span class="rfc2119">MAY</span> be
documented by including a plain text document directly in the storage root, thus making the storage root
self-documenting.</p>

<h3 id="filesystem-features">4.6 Filesystem features<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#filesystem-features" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>In order to maximize the compatibility of the OCFL with different filesystems, and thus improve the portability of OCFL
Objects between different systems, some restrictions on the use of certain filesystem features are necessary. If the
preservation of non-OCFL-compliant features is required then the content <span id="E089" class="rfc2119">MUST</span> be
wrapped in a suitable disk or filesystem image format which OCFL can treat as a regular file.</p>

<ol>
  <li>
    <p>Filesystem metadata (e.g. permissions, access, and creation times) are not considered portable between filesystems or
preservable through file transfer operations. These attributes also cannot be validated in terms of fixity in a
consistent manner. As such, the OCFL does not support the portability of these attributes.</p>
  </li>
  <li>
    <p>Hard and soft (symbolic) links are not portable and <span id="E090" class="rfc2119">MUST NOT</span> be used within
OCFL Storage hierarchies. A common use case for links is storage deduplication. OCFL inventories provide a portable
method of achieving the same effect by using digests to address content.</p>
  </li>
  <li>
    <p>File paths and filenames in the OCFL are case sensitive. Implementations over filesystems that either do not preserve
case or are not case sensitive require great care, including making appropriate choices for file paths and filenames.</p>
  </li>
  <li>
    <p>Transparent filesystem features such as compression and encryption should be effectively invisible to OCFL
operations. Consequently, they should not be expected to be portable.</p>
  </li>
</ol>

<h2 id="examples">5. Examples<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#examples" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<p><em>This section is non-normative.</em></p>

<h3 id="example-minimal-object">5.1 Minimal OCFL Object<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#example-minimal-object" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The following example OCFL Object has content that is a single file (<code class="language-plaintext highlighter-rouge">file.txt</code>), and just one version (<code class="language-plaintext highlighter-rouge">v1</code>):</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[object root]
    ├── 0=ocfl_object_1.1
    ├── inventory.json
    ├── inventory.json.sha512
    └── v1
        ├── inventory.json
        ├── inventory.json.sha512
        └── content
            └── file.txt
</code></pre></div></div>

<p>The inventory for this OCFL Object, the same both at the top-level and in the <code class="language-plaintext highlighter-rouge">v1</code> directory, might be:</p>

<div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="p">{</span><span class="w">
  </span><span class="nl">"digestAlgorithm"</span><span class="p">:</span><span class="w"> </span><span class="s2">"sha512"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"head"</span><span class="p">:</span><span class="w"> </span><span class="s2">"v1"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"id"</span><span class="p">:</span><span class="w"> </span><span class="s2">"http://example.org/minimal"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"manifest"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"7545b8...f67"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/file.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
  </span><span class="p">},</span><span class="w">
  </span><span class="nl">"type"</span><span class="p">:</span><span class="w"> </span><span class="s2">"https://ocfl.io/1.1/spec/#inventory"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"versions"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"v1"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2018-10-02T12:00:00Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"message"</span><span class="p">:</span><span class="w"> </span><span class="s2">"One file"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"7545b8...f67"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"file.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">},</span><span class="w">
      </span><span class="nl">"user"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"address"</span><span class="p">:</span><span class="w"> </span><span class="s2">"mailto:alice@example.org"</span><span class="p">,</span><span class="w">
        </span><span class="nl">"name"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Alice"</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">}</span><span class="w">
  </span><span class="p">}</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div></div>

<h3 id="example-versioned-object">5.2 Versioned OCFL Object<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#example-versioned-object" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The following example OCFL Object has three versions:</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[object root]
    ├── 0=ocfl_object_1.1
    ├── inventory.json
    ├── inventory.json.sha512
    ├── v1
    │&nbsp;&nbsp; ├── inventory.json
    │&nbsp;&nbsp; ├── inventory.json.sha512
    │&nbsp;&nbsp; └── content
    │       ├── empty.txt
    │       ├── foo
    │       │&nbsp;&nbsp; └── bar.xml
    │       └── image.tiff
    ├── v2
    │&nbsp;&nbsp; ├── inventory.json
    │&nbsp;&nbsp; ├── inventory.json.sha512
    │&nbsp;&nbsp; └── content
    │       └── foo
    │       &nbsp; &nbsp; └── bar.xml
    └── v3
        ├── inventory.json
        └── inventory.json.sha512
</code></pre></div></div>

<p>In <code class="language-plaintext highlighter-rouge">v1</code> there are three files, <code class="language-plaintext highlighter-rouge">empty.txt</code>, <code class="language-plaintext highlighter-rouge">foo/bar.xml</code>, and <code class="language-plaintext highlighter-rouge">image.tiff</code>. In <code class="language-plaintext highlighter-rouge">v2</code> the content of <code class="language-plaintext highlighter-rouge">foo/bar.xml</code> is
changed, <code class="language-plaintext highlighter-rouge">empty2.txt</code> is added with the same content as <code class="language-plaintext highlighter-rouge">empty.txt</code>, and <code class="language-plaintext highlighter-rouge">image.tiff</code> is removed. In <code class="language-plaintext highlighter-rouge">v3</code> the file
<code class="language-plaintext highlighter-rouge">empty.txt</code> is removed, and <code class="language-plaintext highlighter-rouge">image.tiff</code> is reinstated. As a result of forward-delta versioning, the object tree above
shows only new content added in each version. The inventory shown below details the other changes, includes additional
fixity information using <code class="language-plaintext highlighter-rouge">md5</code> and <code class="language-plaintext highlighter-rouge">sha1</code> digest algorithms, and minimal metadata for each version.</p>

<div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="p">{</span><span class="w">
  </span><span class="nl">"digestAlgorithm"</span><span class="p">:</span><span class="w"> </span><span class="s2">"sha512"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"fixity"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"md5"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"184f84e28cbe75e050e9c25ea7f2e939"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
      </span><span class="nl">"2673a7b11a70bc7ff960ad8127b4adeb"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v2/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
      </span><span class="nl">"c289c8ccd4bab6e385f5afdd89b5bda2"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/image.tiff"</span><span class="w"> </span><span class="p">],</span><span class="w">
      </span><span class="nl">"d41d8cd98f00b204e9800998ecf8427e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/empty.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
    </span><span class="p">},</span><span class="w">
    </span><span class="nl">"sha1"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"66709b068a2faead97113559db78ccd44712cbf2"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
      </span><span class="nl">"a6357c99ecc5752931e133227581e914968f3b9c"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v2/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
      </span><span class="nl">"b9c7ccc6154974288132b63c15db8d2750716b49"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/image.tiff"</span><span class="w"> </span><span class="p">],</span><span class="w">
      </span><span class="nl">"da39a3ee5e6b4b0d3255bfef95601890afd80709"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/empty.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
    </span><span class="p">}</span><span class="w">
  </span><span class="p">},</span><span class="w">
  </span><span class="nl">"head"</span><span class="p">:</span><span class="w"> </span><span class="s2">"v3"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"id"</span><span class="p">:</span><span class="w"> </span><span class="s2">"ark:/12345/bcd987"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"manifest"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"4d27c8...b53"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v2/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"7dcc35...c31"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/empty.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"ffccf6...62e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/image.tiff"</span><span class="w"> </span><span class="p">]</span><span class="w">
  </span><span class="p">},</span><span class="w">
  </span><span class="nl">"type"</span><span class="p">:</span><span class="w"> </span><span class="s2">"https://ocfl.io/1.1/spec/#inventory"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"versions"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"v1"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2018-01-01T01:01:01Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"message"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Initial import"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"7dcc35...c31"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"empty.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"ffccf6...62e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"image.tiff"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">},</span><span class="w">
      </span><span class="nl">"user"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"address"</span><span class="p">:</span><span class="w"> </span><span class="s2">"mailto:alice@example.com"</span><span class="p">,</span><span class="w">
        </span><span class="nl">"name"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Alice"</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">},</span><span class="w">
    </span><span class="nl">"v2"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2018-02-02T02:02:02Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"message"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Fix bar.xml, remove image.tiff, add empty2.txt"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"4d27c8...b53"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"empty.txt"</span><span class="p">,</span><span class="w"> </span><span class="s2">"empty2.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">},</span><span class="w">
      </span><span class="nl">"user"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"address"</span><span class="p">:</span><span class="w"> </span><span class="s2">"mailto:bob@example.com"</span><span class="p">,</span><span class="w">
        </span><span class="nl">"name"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Bob"</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">},</span><span class="w">
    </span><span class="nl">"v3"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2018-03-03T03:03:03Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"message"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Reinstate image.tiff, delete empty.txt"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"4d27c8...b53"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"foo/bar.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"empty2.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"ffccf6...62e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"image.tiff"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">},</span><span class="w">
      </span><span class="nl">"user"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"address"</span><span class="p">:</span><span class="w"> </span><span class="s2">"mailto:cecilia@example.com"</span><span class="p">,</span><span class="w">
        </span><span class="nl">"name"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Cecilia"</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">}</span><span class="w">
  </span><span class="p">}</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div></div>

<h3 id="example-object-diff-paths">5.3 Different Logical and Content Paths in an OCFL Object<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#example-object-diff-paths" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The following example OCFL Object inventory shows how content paths may differ from logical paths. The example object
has just one version, <code class="language-plaintext highlighter-rouge">v1</code>, which has two files with logical paths <code class="language-plaintext highlighter-rouge">a file.wxy</code> and <code class="language-plaintext highlighter-rouge">another file.xyz</code> as shown in the
<code class="language-plaintext highlighter-rouge">state</code> block. The corresponding content paths are <code class="language-plaintext highlighter-rouge">v1/content/3bacb119a98a15c5</code> and <code class="language-plaintext highlighter-rouge">v1/content/9f2bab8ef869947d</code>
respectively, as shown in the <code class="language-plaintext highlighter-rouge">manifest</code>. Except for location within the appropriate version directory, <code class="language-plaintext highlighter-rouge">v1/content</code> in
this example, the OCFL specification does not constrain the choice of content paths used when creating or updating an
OCFL object. The choice might depend on particular limitations of, or optimizations for, the target storage system, or
on portability considerations. Any compliant implementation will be able to recover version state with the original
logical paths.</p>

<div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="p">{</span><span class="w">
  </span><span class="nl">"digestAlgorithm"</span><span class="p">:</span><span class="w"> </span><span class="s2">"sha512"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"head"</span><span class="p">:</span><span class="w"> </span><span class="s2">"v1"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"id"</span><span class="p">:</span><span class="w"> </span><span class="s2">"http://example.org/diff-paths"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"manifest"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"7545b8...f67"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/3bacb119a98a15c5"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"af318d...3cd"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/9f2bab8ef869947d"</span><span class="w"> </span><span class="p">]</span><span class="w">
  </span><span class="p">},</span><span class="w">
  </span><span class="nl">"type"</span><span class="p">:</span><span class="w"> </span><span class="s2">"https://ocfl.io/1.1/spec/#inventory"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"versions"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"v1"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2019-03-14T20:31:00Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"7545b8...f67"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"a file.wxy"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"af318d...3cd"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"another file.xyz"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">},</span><span class="w">
      </span><span class="nl">"user"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"address"</span><span class="p">:</span><span class="w"> </span><span class="s2">"mailto:admin@example.org"</span><span class="p">,</span><span class="w">
        </span><span class="nl">"name"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Some Admin"</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">}</span><span class="w">
  </span><span class="p">}</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div></div>

<h3 id="example-bagit-in-ocfl">5.4 BagIt in an OCFL Object<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#example-bagit-in-ocfl" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>[<a href="https://ocfl.io/1.1/spec/#ref-bagit">BagIt</a>] is a common file packaging specification, but unlike the OCFL it does not provide a mechanism
for content versioning. Using the OCFL it is possible to store a BagIt structure with content versioning, such that when
the <a href="https://ocfl.io/1.1/spec/#dfn-logical-state">logical state</a> is resolved, it creates a valid BagIt ‘bag’. This example will illustrate one
way this can be accomplished, using the <a href="https://datatracker.ietf.org/doc/html/rfc8493#section-4.1">example of a basic
bag</a> given in the BagIt specification.</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[object root]
    ├── 0=ocfl_object_1.1
    ├── inventory.json
    ├── inventory.json.sha512
    └── v1
        ├── inventory.json
        ├── inventory.json.sha512
        └── content
            └── myfirstbag
                ├── bagit.txt
                ├── data
                │&nbsp;&nbsp; └── 27613-h
                │&nbsp;&nbsp;     └── images
                │&nbsp;&nbsp;         ├── q172.png
                │&nbsp;&nbsp;         └── q172.txt
                └── manifest-md5.txt
</code></pre></div></div>

<p>If, for example, a new directory were added in a subsequent version, the OCFL Object would look like this:</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[object root]
    ├── 0=ocfl_object_1.1
    ├── inventory.json
    ├── inventory.json.sha512
    ├── v1
    │   ├── inventory.json
    │   ├── inventory.json.sha512
    │   └── content
    │&nbsp;&nbsp;     └── myfirstbag
    │&nbsp;&nbsp;         ├── bagit.txt
    │&nbsp;&nbsp;         ├── data
    │&nbsp;&nbsp;         │&nbsp;&nbsp; └── 27613-h
    │&nbsp;&nbsp;         │&nbsp;&nbsp;     └── images
    │&nbsp;&nbsp;         │&nbsp;&nbsp;         ├── q172.png
    │&nbsp;&nbsp;         │&nbsp;&nbsp;         └── q172.txt
    │&nbsp;&nbsp;         └── manifest-md5.txt
    └── v2
        ├── inventory.json
        ├── inventory.json.sha512
        └── content
            └── myfirstbag
                ├── data
                │&nbsp;&nbsp; └── 27614-h
                │&nbsp;&nbsp;     └── images
                │&nbsp;&nbsp;         ├── q173.png
                │&nbsp;&nbsp;         └── q173.txt
                └── manifest-md5.txt
</code></pre></div></div>

<p>The state of the object at version 2 would be the following BagIt object:</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>myfirstbag
    ├── bagit.txt
    ├── data
    │&nbsp;&nbsp; ├── 27613-h
    │&nbsp;&nbsp; │&nbsp;&nbsp; └── images
    │&nbsp;&nbsp; │&nbsp;&nbsp;     ├── q172.png
    │&nbsp;&nbsp; │&nbsp;&nbsp;     └── q172.txt
    │&nbsp;&nbsp; └── 27614-h
    │&nbsp;&nbsp;     └── images
    │&nbsp;&nbsp;         ├── q173.png
    │&nbsp;&nbsp;         └── q173.txt
    └── manifest-md5.txt
</code></pre></div></div>

<p>The OCFL Inventory for this object would be as follows:</p>

<div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="p">{</span><span class="w">
  </span><span class="nl">"digestAlgorithm"</span><span class="p">:</span><span class="w"> </span><span class="s2">"sha512"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"head"</span><span class="p">:</span><span class="w"> </span><span class="s2">"v2"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"id"</span><span class="p">:</span><span class="w"> </span><span class="s2">"urn:uri:example.com/myfirstbag"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"manifest"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/myfirstbag/bagit.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"f15428...83f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/myfirstbag/manifest-md5.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"85f2b0...007"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/myfirstbag/data/27613-h/images/q172.png"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"d66d80...8bd"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v1/content/myfirstbag/data/27613-h/images/q172.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"2b0ff8...620"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v2/content/myfirstbag/manifest-md5.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"921d36...877"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v2/content/myfirstbag/data/27614-h/images/q173.png"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"b8bdf1...927"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v2/content/myfirstbag/data/27614-h/images/q173.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
  </span><span class="p">},</span><span class="w">
  </span><span class="nl">"type"</span><span class="p">:</span><span class="w"> </span><span class="s2">"https://ocfl.io/1.1/spec/#inventory"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"versions"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"v1"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2018-10-09T11:20:29.209164Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"message"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Initial Ingest"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/bagit.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"85f2b0...007"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/data/27613-h/images/q172.png"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"d66d80...8bd"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/data/27613-h/images/q172.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"f15428...83f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/manifest-md5.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">},</span><span class="w">
      </span><span class="nl">"user"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"address"</span><span class="p">:</span><span class="w"> </span><span class="s2">"mailto:someone@example.org"</span><span class="p">,</span><span class="w">
        </span><span class="nl">"name"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Some One"</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">},</span><span class="w">
    </span><span class="nl">"v2"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2018-10-31T11:20:29.209164Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"message"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Added new images"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"cf83e1...a3e"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/bagit.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"85f2b0...007"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/data/27613-h/images/q172.png"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"d66d80...8bd"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/data/27613-h/images/q172.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"2b0ff8...620"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/manifest-md5.txt"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"921d36...877"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/data/27614-h/images/q173.png"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"b8bdf1...927"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"myfirstbag/data/27614-h/images/q173.txt"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">},</span><span class="w">
      </span><span class="nl">"user"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"address"</span><span class="p">:</span><span class="w"> </span><span class="s2">"mailto:somebody-else@example.org"</span><span class="p">,</span><span class="w">
        </span><span class="nl">"name"</span><span class="p">:</span><span class="w"> </span><span class="s2">"Somebody Else"</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">}</span><span class="w">
  </span><span class="p">}</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div></div>

<h3 id="example-moab-in-ocfl">5.5 Moab in an OCFL Object<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#example-moab-in-ocfl" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>[<a href="https://ocfl.io/1.1/spec/#ref-moab">Moab</a>] is an archive information package format developed and used by Stanford University. Many of the
ideas in Moab have been refined by the OCFL, and the OCFL is designed to give institutions currently using Moab an easy
path to adoption.</p>

<p>Converting content preserved in a Moab object in a way that does not compromise existing Moab access patterns whilst
allowing for the eventual use of OCFL-native workflows requires a Moab to OCFL conversion tool. This tool uses the
Moab-versioning gem to extract deltas and digests of the Moab data directory for each Moab version and translate those
into version <code class="language-plaintext highlighter-rouge">state</code> blocks in an OCFL inventory file, which would be placed in the root directory of the Moab object.
The content of the <code class="language-plaintext highlighter-rouge">data</code> directory in the Moab version directories (and thus, the bitstreams that Moab is preserving)
is tracked by OCFL, via the <code class="language-plaintext highlighter-rouge">contentDirectory</code> value. The contents of the Moab <code class="language-plaintext highlighter-rouge">manifests</code> directories are not tracked,
as the intention is not to encapsulate a Moab object inside an OCFL object, but rather to migrate Moab’s preserved
bitstreams into an OCFL object without compromising legacy access patterns.</p>

<p>During the transitionary period the OCFL inventory file exists only in the root of the Moab object. Once OCFL-native
object creation workflows have been completed, future versions of that object will be fully OCFL compliant - new
versions will no longer have a manifests directory and will contain an OCFL inventory file. At this stage OCFL tools
will be able to access all versions of the content originally preserved by Moab.</p>

<p>Consider the following sample Moab object:</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[object root]
    └── bj102hs9687
        ├── v0001
        │&nbsp;&nbsp;   ├── data
        │&nbsp;&nbsp;   │&nbsp;&nbsp; ├── content
        │&nbsp;&nbsp;   │&nbsp;&nbsp; │&nbsp;&nbsp; ├── eric-smith-dissertation-augmented.pdf
        │&nbsp;&nbsp;   │&nbsp;&nbsp; │&nbsp;&nbsp; └── eric-smith-dissertation.pdf
        │&nbsp;&nbsp;   │&nbsp;&nbsp; └── metadata
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     ├── contentMetadata.xml
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     ├── descMetadata.xml
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     ├── identityMetadata.xml
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     ├── provenanceMetadata.xml
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     ├── relationshipMetadata.xml
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     ├── rightsMetadata.xml
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     ├── technicalMetadata.xml
        │&nbsp;&nbsp;   │&nbsp;&nbsp;     └── versionMetadata.xml
        │&nbsp;&nbsp;   └── manifests
        │&nbsp;&nbsp;       ├── fileInventoryDifference.xml
        │&nbsp;&nbsp;       ├── manifestInventory.xml
        │&nbsp;&nbsp;       ├── signatureCatalog.xml
        │&nbsp;&nbsp;       ├── versionAdditions.xml
        │&nbsp;&nbsp;       └── versionInventory.xml
        ├── v0002
        │&nbsp;&nbsp;   ├── data
        │&nbsp;&nbsp;   │&nbsp;&nbsp; └── metadata
        │&nbsp;    │&nbsp;&nbsp;     ├── contentMetadata.xml
        │&nbsp;    │&nbsp;&nbsp;     ├── embargoMetadata.xml
        │&nbsp;    │&nbsp;&nbsp;     ├── events.xml
        │&nbsp;    │&nbsp;&nbsp;     ├── identityMetadata.xml
        │&nbsp;    │&nbsp;&nbsp;     ├── provenanceMetadata.xml
        │&nbsp;    │&nbsp;&nbsp;     ├── relationshipMetadata.xml
        │&nbsp;    │&nbsp;&nbsp;     ├── rightsMetadata.xml
        │&nbsp;    │&nbsp;&nbsp;     ├── versionMetadata.xml
        │&nbsp;    │&nbsp;&nbsp;     └── workflows.xml
        │&nbsp;    └── manifests
        │&nbsp;        ├── fileInventoryDifference.xml
        │&nbsp;        ├── manifestInventory.xml
        │&nbsp;        ├── signatureCatalog.xml
        │&nbsp;        ├── versionAdditions.xml
        │&nbsp;        └── versionInventory.xml
        └── v0003
              ├── data
              │&nbsp;&nbsp; └── metadata
              │&nbsp;&nbsp;     ├── contentMetadata.xml
              │&nbsp;&nbsp;     ├── descMetadata.xml
              │&nbsp;&nbsp;     ├── embargoMetadata.xml
              │&nbsp;&nbsp;     ├── events.xml
              │&nbsp;&nbsp;     ├── identityMetadata.xml
              │&nbsp;&nbsp;     ├── provenanceMetadata.xml
              │&nbsp;&nbsp;     ├── rightsMetadata.xml
              │&nbsp;&nbsp;     ├── technicalMetadata.xml
              │&nbsp;&nbsp;     ├── versionMetadata.xml
              │&nbsp;&nbsp;     └── workflows.xml
              └── manifests
                  ├── fileInventoryDifference.xml
                  ├── manifestInventory.xml
                  ├── signatureCatalog.xml
                  ├── versionAdditions.xml
                  └── versionInventory.xml
</code></pre></div></div>

<p>An OCFL inventory that tracks the <code class="language-plaintext highlighter-rouge">data</code> directory would include a manifest comprised as follows. Note the absence of
the <code class="language-plaintext highlighter-rouge">manifests</code> directory, as we are not encapsulating the Moab object in an OCFL object, and the presence of
<code class="language-plaintext highlighter-rouge">contentDirectory</code> to specify <code class="language-plaintext highlighter-rouge">data</code> as the preserved content directory:</p>

<div class="language-json highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="p">{</span><span class="w">
  </span><span class="nl">"digestAlgorithm"</span><span class="p">:</span><span class="w"> </span><span class="s2">"sha512"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"head"</span><span class="p">:</span><span class="w"> </span><span class="s2">"v3"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"id"</span><span class="p">:</span><span class="w"> </span><span class="s2">"druid:bj102hs9687"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"contentDirectory"</span><span class="p">:</span><span class="w"> </span><span class="s2">"data"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"manifest"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"98114a...588"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/content/eric-smith-dissertation-augmented.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"7f3d87...15b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/content/eric-smith-dissertation.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"6d19f0...064"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/technicalMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"6e4be4...375"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/provenanceMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"d8a319...d0f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/descMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"de823a...acc"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/rightsMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"080617...40c"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/identityMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"e15267...58d"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/versionMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"0d9e0b...9a2"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/contentMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"dd9289...31d"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0001/data/metadata/relationshipMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"7519c5...63f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/provenanceMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"abda4c...622"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/workflows.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"76549e...b2b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/rightsMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"bdc4d6...3b6"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/events.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"7b331c...f9b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/identityMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"80ceac...b9c"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/versionMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"4853a2...fbe"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/contentMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"1d5090...f5f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/relationshipMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"f209bf...ceb"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0002/data/metadata/embargoMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"dd9125...d4b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/technicalMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"d9e177...477"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/provenanceMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"4f5908...4f5"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/workflows.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"e64db0...500"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/descMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"05fa51...818"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/rightsMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"d70dd8...5ad"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/events.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"509a2d...dc6"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/identityMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"548066...893"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/versionMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"93884e...aae"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/contentMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
    </span><span class="nl">"4c5ab4...b02"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"v0003/data/metadata/embargoMetadata.xml"</span><span class="w"> </span><span class="p">]</span><span class="w">
  </span><span class="p">},</span><span class="w">
  </span><span class="nl">"type"</span><span class="p">:</span><span class="w"> </span><span class="s2">"https://ocfl.io/1.1/spec/#inventory"</span><span class="p">,</span><span class="w">
  </span><span class="nl">"versions"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
    </span><span class="nl">"v1"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2019-03-14T20:31:00Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"98114a...588"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"content/eric-smith-dissertation-augmented.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"7f3d87...15b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"content/eric-smith-dissertation.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"6d19f0...064"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/technicalMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"6e4be4...375"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/provenanceMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"d8a319...d0f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/descMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"de823a...acc"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/rightsMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"080617...40c"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/identityMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"e15267...58d"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/versionMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"0d9e0b...9a2"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/contentMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"dd9289...31d"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/relationshipMetadata.xml"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">},</span><span class="w">
    </span><span class="nl">"v2"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2019-03-24T09:22:00Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"98114a...588"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"content/eric-smith-dissertation-augmented.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"7f3d87...15b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"content/eric-smith-dissertation.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"6d19f0...064"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/technicalMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"7519c5...63f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/provenanceMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"d8a319...d0f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/descMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"76549e...b2b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/rightsMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"7b331c...f9b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/identityMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"80ceac...b9c"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/versionMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"4853a2...fbe"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/contentMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"1d5090...f5f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/relationshipMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"abda4c...622"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/workflows.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"bdc4d6...3b6"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/events.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"f209bf...ceb"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/embargoMetadata.xml"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">},</span><span class="w">
    </span><span class="nl">"v3"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
      </span><span class="nl">"created"</span><span class="p">:</span><span class="w"> </span><span class="s2">"2019-04-02T11:07:00Z"</span><span class="p">,</span><span class="w">
      </span><span class="nl">"state"</span><span class="p">:</span><span class="w"> </span><span class="p">{</span><span class="w">
        </span><span class="nl">"98114a...588"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"content/eric-smith-dissertation-augmented.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"7f3d87...15b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"content/eric-smith-dissertation.pdf"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"dd9125...d4b"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/technicalMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"d9e177...477"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/provenanceMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"e64db0...500"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/descMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"05fa51...818"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/rightsMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"509a2d...dc6"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/identityMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"548066...893"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/versionMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"93884e...aae"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/contentMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"1d5090...f5f"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/relationshipMetadata.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"4f5908...4f5"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/workflows.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"d70dd8...5ad"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/events.xml"</span><span class="w"> </span><span class="p">],</span><span class="w">
        </span><span class="nl">"4c5ab4...b02"</span><span class="p">:</span><span class="w"> </span><span class="p">[</span><span class="w"> </span><span class="s2">"metadata/embargoMetadata.xml"</span><span class="w"> </span><span class="p">]</span><span class="w">
      </span><span class="p">}</span><span class="w">
    </span><span class="p">}</span><span class="w">
  </span><span class="p">}</span><span class="w">
</span><span class="p">}</span><span class="w">
</span></code></pre></div></div>

<h3 id="example-extended-storage-root">5.6 Example Extended OCFL Storage Root<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#example-extended-storage-root" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The following example OCFL Storage Root has an extension containing custom content. The OCFL Storage Root itself remains
valid.</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[storage root]
    ├── 0=ocfl_1.1
    ├── extensions
    │&nbsp;&nbsp; └── 0000-example-extension
    │&nbsp;&nbsp;     └── file-example.txt
    ├── ocfl_1.1.txt
    └── ocfl_layout.json
</code></pre></div></div>

<h3 id="example-extended-object">5.7 Example Extended OCFL Object<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#example-extended-object" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p>The following example OCFL Object has an extension containing custom content. The OCFL Object itself remains valid.</p>

<div class="language-plaintext highlighter-rouge"><div class="highlight"><pre class="highlight"><code>[object root]
    ├── 0=ocfl_object_1.1
    ├── inventory.json
    ├── inventory.json.sha512
    ├── extensions
    │&nbsp;&nbsp; └── 0000-example-extension
    │&nbsp;&nbsp;     └── file1-draft.txt
    └── v1
        ├── inventory.json
        ├── inventory.json.sha512
        └── content
            └── file.txt
</code></pre></div></div>

<h2 id="references">6. References<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#references" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<h3 id="normative-references">6.1 Normative References<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#normative-references" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p><span id="ref-fips-180-4"></span><strong>[FIPS-180-4]</strong> FIPS PUB 180-4 Secure Hash Standard. U.S. Department of Commerce/National
Institute of Standards and Technology. URL: <a href="https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf">https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf</a></p>

<p><span id="ref-namaste"></span><strong>[NAMASTE]</strong> Directory Description with Namaste Tags. J. Kunze. 9 November 2009. URL:
<a href="https://n2t.net/ark:/13030/c7g44hq41">https://n2t.net/ark:/13030/c7g44hq41</a>, local copy: <a href="https://ocfl.io/cache/NamasteSpec.pdf">https://ocfl.io/cache/NamasteSpec.pdf</a></p>

<p><span id="ref-rfc1321"></span><strong>[RFC1321]</strong> The MD5 Message-Digest Algorithm. R. Rivest. IETF. April 1992. Informational.
URL: <a href="https://www.rfc-editor.org/rfc/rfc1321">https://www.rfc-editor.org/rfc/rfc1321</a></p>

<p><span id="ref-rfc2119"></span><strong>[RFC2119]</strong> Key words for use in RFCs to Indicate Requirement Levels. S. Bradner. IETF.
March 1997. Best Current Practice. URL: <a href="https://www.rfc-editor.org/rfc/rfc2119">https://www.rfc-editor.org/rfc/rfc2119</a></p>

<p><span id="ref-rfc3339"></span><strong>[RFC3339]</strong> Date and Time on the Internet: Timestamps. G. Klyne; C. Newman. IETF. July 2002.
Proposed Standard. URL: <a href="https://www.rfc-editor.org/rfc/rfc3339">https://www.rfc-editor.org/rfc/rfc3339</a></p>

<p><span id="ref-rfc3986"></span><strong>[RFC3986]</strong> Uniform Resource Identifier (URI): Generic Syntax. T. Berners-Lee; R. Fielding;
L. Masinter. IETF. January 2005. Internet Standard. URL: <a href="https://www.rfc-editor.org/rfc/rfc3986">https://www.rfc-editor.org/rfc/rfc3986</a></p>

<p><span id="ref-rfc4648"></span><strong>[RFC4648]</strong> The Base16, Base32, and Base64 Data Encodings. S. Josefsson. IETF. October 2006.
Proposed Standard. URL: <a href="https://www.rfc-editor.org/rfc/rfc4648">https://www.rfc-editor.org/rfc/rfc4648</a></p>

<p><span id="ref-rfc7693"></span><strong>[RFC7693]</strong> The BLAKE2 Cryptographic Hash and Message Authentication Code (MAC). M-J.
Saarinen, Ed.; J-P. Aumasson. IETF. November 2015. Informational. URL: <a href="https://www.rfc-editor.org/rfc/rfc7693">https://www.rfc-editor.org/rfc/rfc7693</a></p>

<p><span id="ref-rfc8259"></span><strong>[RFC8259]</strong> The JavaScript Object Notation (JSON) Data Interchange Format. T. Bray, Ed..
IETF. December 2017. Internet Standard. URL: <a href="https://www.rfc-editor.org/rfc/rfc8259">https://www.rfc-editor.org/rfc/rfc8259</a></p>

<h3 id="informative-references">6.2 Informative References<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#informative-references" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h3>

<p><span id="ref-bagit"></span><strong>[BagIt]</strong> The BagIt File Packaging Format (V1.0). J. Kunze; J. Littman; E. Madden; J.
Scancella; C. Adams. 17 September 2018. URL: <a href="https://datatracker.ietf.org/doc/html/rfc8493">https://datatracker.ietf.org/doc/html/rfc8493</a></p>

<p><span id="ref-digest-algorithms-extension"></span><strong>[Digest-Algorithms-Extension]</strong> OCFL Community Extension 0001: Digest
Algorithms. OCFL Editors.URL: <a href="https://ocfl.github.io/extensions/0001-digest-algorithms.html">https://ocfl.github.io/extensions/0001-digest-algorithms.html</a></p>

<p><span id="ref-json-schema"></span><strong>[JSON-Schema]</strong> JSON Schema Validation: A Vocabulary for Structural Validation of JSON.
A. Wright; H Andrews.20 September 2018. URL: <a href="https://json-schema.org/latest/json-schema-validation.html">https://json-schema.org/latest/json-schema-validation.html</a></p>

<p><span id="ref-moab"></span><strong>[Moab]</strong> The Moab Design for Digital Object Versioning. Richard Anderson.15 July 2013. URL:
<a href="https://journal.code4lib.org/articles/8482">https://journal.code4lib.org/articles/8482</a></p>

<p><span id="ref-oais"></span><strong>[OAIS]</strong> Reference Model for an Open Archival Information System (OAIS), Issue 2. June 2012.
URL: <a href="https://public.ccsds.org/pubs/650x0m2.pdf">https://public.ccsds.org/pubs/650x0m2.pdf</a></p>

<p><span id="ref-ocfl-implementation-notes"></span><strong>[OCFL-Implementation-Notes]</strong> OCFL Implementation Notes. URL:
<a href="https://ocfl.io/1.1/implementation-notes">https://ocfl.io/1.1/implementation-notes</a></p>

<p><span id="ref-pairtree"></span><strong>[PairTree]</strong> Pairtrees for Object Storage. J. Kunze; M. Haye; E. Hetzner; M. Reyes; C.
Snavely. 12 August 2008. URL: <a href="https://datatracker.ietf.org/doc/html/draft-kunze-pairtree-01">https://datatracker.ietf.org/doc/html/draft-kunze-pairtree-01</a></p>

<p><span id="ref-rfc6068"></span><strong>[RFC6068]</strong> The ‘mailto’ URI Scheme. M. Duerst; L. Masinter; J. Zawinski. IETF. October 2010.
Proposed Standard. URL: <a href="https://www.rfc-editor.org/rfc/rfc6068">https://www.rfc-editor.org/rfc/rfc6068</a></p>

<p><span id="ref-rfc7764"></span><strong>[RFC7764]</strong> Guidance on Markdown: Design Philosophies, Stability Strategies, and Select
Registrations. S. Leonard. IETF. March 2016. URL:  <a href="https://www.rfc-editor.org/rfc/rfc7764">https://www.rfc-editor.org/rfc/rfc7764</a></p>

<p><span id="ref-rfc8141"></span><strong>[RFC8141]</strong> Uniform Resource Names (URNs). P. Saint-Andre; J. Klensin. IETF. April 2017.
Proposed Standard. URL: <a href="https://www.rfc-editor.org/rfc/rfc8141">https://www.rfc-editor.org/rfc/rfc8141</a></p>

<h2 id="revision-history">Revision history<a class="anchorjs-link " href="https://ocfl.io/1.1/spec/#revision-history" aria-label="Anchor" data-anchorjs-icon="" style="font: 1em / 1 anchorjs-icons; padding-left: 0.375em;"></a></h2>

<table>
  <thead>
    <tr>
      <th>Version</th>
      <th>Date</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><a href="https://ocfl.io/1.1.0/spec/">v1.1.0</a></td>
      <td>7 October 2022</td>
      <td>First v1.1, <a href="https://ocfl.io/news/#version-11-of-the-oxford-common-file-layout-ocfl-released">release notes</a>, <a href="https://ocfl.io/1.1.0/spec/change-log.html">change log</a></td>
    </tr>
    <tr>
      <td><a href="https://ocfl.io/1.1.1/spec/">v1.1.1</a></td>
      <td>7 November 2024</td>
      <td>Clarifications, <a href="https://ocfl.io/news/#version-111-of-the-oxford-common-file-layout-ocfl-released">release notes</a>, <a href="https://ocfl.io/1.1.1/spec/change-log.html">change log</a></td>
    </tr>
  </tbody>
</table>


    </div>
    <script src="./ocfl11_files/anchor.min.js.Download" integrity="sha256-lZaRhKri35AyJSypXXs4o6OPFTbTmUoltBbDCbdzegg=" crossorigin="anonymous"></script>
    <script>anchors.add();</script>
  

<div id="gridmanWrapper" style="position: fixed; height: 0px; width: 0px; z-index: 99999; top: 0px; left: 0px;"><template shadowrootmode="open"><style>#gridmanContainer blockquote,#gridmanContainer dl,#gridmanContainer dd,#gridmanContainer h1,#gridmanContainer h2,#gridmanContainer h3,#gridmanContainer h4,#gridmanContainer h5,#gridmanContainer h6,#gridmanContainer hr,#gridmanContainer figure,#gridmanContainer p,#gridmanContainer pre{margin:0}#gridmanContainer h1,#gridmanContainer h2,#gridmanContainer h3,#gridmanContainer h4,#gridmanContainer h5,#gridmanContainer h6{font-size:inherit;font-weight:inherit}#gridmanContainer ol,#gridmanContainer ul{list-style:none;margin:0;padding:0}#gridmanContainer img,#gridmanContainer svg,#gridmanContainer video,#gridmanContainer canvas,#gridmanContainer audio,#gridmanContainer iframe,#gridmanContainer embed,#gridmanContainer object{display:block;vertical-align:middle}#gridmanContainer img,#gridmanContainer video{max-width:100%;height:auto}#gridmanContainer a{cursor:pointer;--tw-text-opacity: 1;color:rgb(255 228 230 / var(--tw-text-opacity))}#gridmanContainer a:hover{text-decoration-line:underline}#gridmanContainer *,#gridmanContainer :before,#gridmanContainer :after{box-sizing:border-box;border-width:0;border-style:solid;border-color:#e5e7eb}#gridmanContainer *{font-family:SF Pro Display;font-size:16px;line-height:1.25;-webkit-tap-highlight-color:rgba(0,0,0,0)}#gridmanContainer *:focus{outline:none}#gridmanContainer *{outline:none}#gridmanContainer *::-webkit-scrollbar{display:none}*,:before,:after{--tw-border-spacing-x: 0;--tw-border-spacing-y: 0;--tw-translate-x: 0;--tw-translate-y: 0;--tw-rotate: 0;--tw-skew-x: 0;--tw-skew-y: 0;--tw-scale-x: 1;--tw-scale-y: 1;--tw-pan-x: ;--tw-pan-y: ;--tw-pinch-zoom: ;--tw-scroll-snap-strictness: proximity;--tw-gradient-from-position: ;--tw-gradient-via-position: ;--tw-gradient-to-position: ;--tw-ordinal: ;--tw-slashed-zero: ;--tw-numeric-figure: ;--tw-numeric-spacing: ;--tw-numeric-fraction: ;--tw-ring-inset: ;--tw-ring-offset-width: 0px;--tw-ring-offset-color: #fff;--tw-ring-color: rgb(59 130 246 / .5);--tw-ring-offset-shadow: 0 0 #0000;--tw-ring-shadow: 0 0 #0000;--tw-shadow: 0 0 #0000;--tw-shadow-colored: 0 0 #0000;--tw-blur: ;--tw-brightness: ;--tw-contrast: ;--tw-grayscale: ;--tw-hue-rotate: ;--tw-invert: ;--tw-saturate: ;--tw-sepia: ;--tw-drop-shadow: ;--tw-backdrop-blur: ;--tw-backdrop-brightness: ;--tw-backdrop-contrast: ;--tw-backdrop-grayscale: ;--tw-backdrop-hue-rotate: ;--tw-backdrop-invert: ;--tw-backdrop-opacity: ;--tw-backdrop-saturate: ;--tw-backdrop-sepia: }::backdrop{--tw-border-spacing-x: 0;--tw-border-spacing-y: 0;--tw-translate-x: 0;--tw-translate-y: 0;--tw-rotate: 0;--tw-skew-x: 0;--tw-skew-y: 0;--tw-scale-x: 1;--tw-scale-y: 1;--tw-pan-x: ;--tw-pan-y: ;--tw-pinch-zoom: ;--tw-scroll-snap-strictness: proximity;--tw-gradient-from-position: ;--tw-gradient-via-position: ;--tw-gradient-to-position: ;--tw-ordinal: ;--tw-slashed-zero: ;--tw-numeric-figure: ;--tw-numeric-spacing: ;--tw-numeric-fraction: ;--tw-ring-inset: ;--tw-ring-offset-width: 0px;--tw-ring-offset-color: #fff;--tw-ring-color: rgb(59 130 246 / .5);--tw-ring-offset-shadow: 0 0 #0000;--tw-ring-shadow: 0 0 #0000;--tw-shadow: 0 0 #0000;--tw-shadow-colored: 0 0 #0000;--tw-blur: ;--tw-brightness: ;--tw-contrast: ;--tw-grayscale: ;--tw-hue-rotate: ;--tw-invert: ;--tw-saturate: ;--tw-sepia: ;--tw-drop-shadow: ;--tw-backdrop-blur: ;--tw-backdrop-brightness: ;--tw-backdrop-contrast: ;--tw-backdrop-grayscale: ;--tw-backdrop-hue-rotate: ;--tw-backdrop-invert: ;--tw-backdrop-opacity: ;--tw-backdrop-saturate: ;--tw-backdrop-sepia: }.container{width:100%}@media (min-width: 640px){.container{max-width:640px}}@media (min-width: 768px){.container{max-width:768px}}@media (min-width: 1024px){.container{max-width:1024px}}@media (min-width: 1280px){.container{max-width:1280px}}@media (min-width: 1536px){.container{max-width:1536px}}.pointer-events-none{pointer-events:none!important}.pointer-events-auto{pointer-events:auto!important}.fixed{position:fixed!important}.absolute{position:absolute!important}.relative{position:relative!important}.-bottom-1{bottom:-4px!important}.-bottom-1\.5{bottom:-6px!important}.-bottom-\[1px\]{bottom:-1px!important}.-bottom-full{bottom:-100%!important}.-left-1{left:-4px!important}.-left-1\.5{left:-6px!important}.-left-5{left:-20px!important}.-left-\[1px\]{left:-1px!important}.-left-full{left:-100%!important}.-right-1{right:-4px!important}.-right-1\.5{right:-6px!important}.-right-full{right:-100%!important}.-top-1{top:-4px!important}.-top-1\.5{top:-6px!important}.-top-5{top:-20px!important}.-top-\[1px\],.-top-px{top:-1px!important}.bottom-0{bottom:0!important}.bottom-1\/3{bottom:33.333333%!important}.left-0{left:0!important}.left-1\/2{left:50%!important}.left-4{left:16px!important}.left-px{left:1px!important}.right-0{right:0!important}.right-2{right:8px!important}.right-4{right:16px!important}.top-0{top:0!important}.top-1\/2{top:50%!important}.top-4{top:16px!important}.top-px{top:1px!important}.z-\[999999\]{z-index:999999!important}.z-\[99999\]{z-index:99999!important}.m-auto{margin:auto!important}.mr-1{margin-right:4px!important}.mt-1{margin-top:4px!important}.mt-3{margin-top:12px!important}.mt-px{margin-top:1px!important}.block{display:block!important}.inline{display:inline!important}.flex{display:flex!important}.grid{display:grid!important}.h-0{height:0!important}.h-3{height:12px!important}.h-4{height:16px!important}.h-5{height:20px!important}.h-6{height:24px!important}.h-8{height:32px!important}.h-9{height:36px!important}.h-\[200px\]{height:200px!important}.h-fit{height:-moz-fit-content!important;height:fit-content!important}.h-full{height:100%!important}.h-px{height:1px!important}.max-h-0{max-height:0!important}.max-h-full{max-height:100%!important}.min-h-2{min-height:8px!important}.w-0{width:0!important}.w-3{width:12px!important}.w-4{width:16px!important}.w-40{width:160px!important}.w-5{width:20px!important}.w-6{width:24px!important}.w-\[200px\]{width:200px!important}.w-fit{width:-moz-fit-content!important;width:fit-content!important}.w-full{width:100%!important}.w-min{width:-moz-min-content!important;width:min-content!important}.min-w-2{min-width:8px!important}.max-w-0{max-width:0!important}.max-w-full{max-width:100%!important}.max-w-screen-sm{max-width:640px!important}.border-collapse{border-collapse:collapse!important}.-translate-x-1\/2{--tw-translate-x: -50% !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.-translate-y-1\/2{--tw-translate-y: -50% !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.translate-x-1\/2{--tw-translate-x: 50% !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.translate-y-1\/2{--tw-translate-y: 50% !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.-rotate-90{--tw-rotate: -90deg !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.rotate-180{--tw-rotate: 180deg !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.rotate-90{--tw-rotate: 90deg !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.transform{transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}@keyframes spin{to{transform:rotate(360deg)}}.animate-spin{animation:spin 1s linear infinite!important}.cursor-not-allowed{cursor:not-allowed!important}.cursor-pointer{cursor:pointer!important}.resize-none{resize:none!important}.resize{resize:both!important}.flex-row{flex-direction:row!important}.flex-col{flex-direction:column!important}.flex-wrap{flex-wrap:wrap!important}.items-start{align-items:flex-start!important}.items-end{align-items:flex-end!important}.items-center{align-items:center!important}.justify-center{justify-content:center!important}.gap-0{gap:0!important}.gap-0\.5{gap:2px!important}.gap-1{gap:4px!important}.gap-2{gap:8px!important}.gap-4{gap:16px!important}.gap-6{gap:24px!important}.gap-\[2px\]{gap:2px!important}.self-start{align-self:flex-start!important}.overflow-hidden{overflow:hidden!important}.whitespace-nowrap{white-space:nowrap!important}.rounded{border-radius:4px!important}.rounded-full{border-radius:9999px!important}.rounded-lg{border-radius:8px!important}.rounded-sm{border-radius:2px!important}.border{border-width:1px!important}.border-0{border-width:0px!important}.border-4{border-width:4px!important}.border-b-2{border-bottom-width:2px!important}.border-l-2{border-left-width:2px!important}.border-r-2{border-right-width:2px!important}.border-t-2{border-top-width:2px!important}.border-solid{border-style:solid!important}.border-rose-200{--tw-border-opacity: 1 !important;border-color:rgb(254 205 211 / var(--tw-border-opacity))!important}.border-rose-300{--tw-border-opacity: 1 !important;border-color:rgb(253 164 175 / var(--tw-border-opacity))!important}.border-rose-500{--tw-border-opacity: 1 !important;border-color:rgb(244 63 94 / var(--tw-border-opacity))!important}.border-rose-700\/30{border-color:#be123c4d!important}.bg-red-500{--tw-bg-opacity: 1 !important;background-color:rgb(239 68 68 / var(--tw-bg-opacity))!important}.bg-rose-400{--tw-bg-opacity: 1 !important;background-color:rgb(251 113 133 / var(--tw-bg-opacity))!important}.bg-rose-500{--tw-bg-opacity: 1 !important;background-color:rgb(244 63 94 / var(--tw-bg-opacity))!important}.bg-rose-500\/10{background-color:#f43f5e1a!important}.bg-rose-500\/5{background-color:#f43f5e0d!important}.bg-rose-500\/50{background-color:#f43f5e80!important}.bg-rose-500\/70{background-color:#f43f5eb3!important}.bg-rose-600{--tw-bg-opacity: 1 !important;background-color:rgb(225 29 72 / var(--tw-bg-opacity))!important}.bg-rose-700{--tw-bg-opacity: 1 !important;background-color:rgb(190 18 60 / var(--tw-bg-opacity))!important}.bg-rose-700\/50{background-color:#be123c80!important}.bg-slate-200{--tw-bg-opacity: 1 !important;background-color:rgb(226 232 240 / var(--tw-bg-opacity))!important}.bg-slate-700{--tw-bg-opacity: 1 !important;background-color:rgb(51 65 85 / var(--tw-bg-opacity))!important}.bg-slate-800{--tw-bg-opacity: 1 !important;background-color:rgb(30 41 59 / var(--tw-bg-opacity))!important}.bg-transparent{background-color:transparent!important}.p-0{padding:0!important}.p-0\.5{padding:2px!important}.p-1{padding:4px!important}.p-1\.5{padding:6px!important}.p-10{padding:40px!important}.p-2{padding:8px!important}.p-4{padding:16px!important}.px-1{padding-left:4px!important;padding-right:4px!important}.px-2{padding-left:8px!important;padding-right:8px!important}.px-3{padding-left:12px!important;padding-right:12px!important}.px-4{padding-left:16px!important;padding-right:16px!important}.py-0{padding-top:0!important;padding-bottom:0!important}.py-1{padding-top:4px!important;padding-bottom:4px!important}.py-2{padding-top:8px!important;padding-bottom:8px!important}.py-px{padding-top:1px!important;padding-bottom:1px!important}.pb-1{padding-bottom:4px!important}.pb-2{padding-bottom:8px!important}.pl-1{padding-left:4px!important}.pl-2{padding-left:8px!important}.pr-1{padding-right:4px!important}.pr-1\.5{padding-right:6px!important}.pr-2{padding-right:8px!important}.pt-0{padding-top:0!important}.pt-1{padding-top:4px!important}.pt-2{padding-top:8px!important}.text-center{text-align:center!important}.text-right{text-align:right!important}.font-mono{font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,Liberation Mono,Courier New,monospace!important}.font-sans{font-family:ui-sans-serif,system-ui,sans-serif,"Apple Color Emoji","Segoe UI Emoji",Segoe UI Symbol,"Noto Color Emoji"!important}.text-2xs{font-size:10px!important}.text-lg{font-size:18px!important}.text-sm{font-size:14px!important}.text-xs{font-size:12px!important}.font-bold{font-weight:700!important}.uppercase{text-transform:uppercase!important}.capitalize{text-transform:capitalize!important}.italic{font-style:italic!important}.leading-normal{line-height:1.5!important}.leading-tight{line-height:1.25!important}.tracking-wide{letter-spacing:.025em!important}.tracking-wider{letter-spacing:.05em!important}.text-rose-200{--tw-text-opacity: 1 !important;color:rgb(254 205 211 / var(--tw-text-opacity))!important}.text-rose-50{--tw-text-opacity: 1 !important;color:rgb(255 241 242 / var(--tw-text-opacity))!important}.text-rose-700{--tw-text-opacity: 1 !important;color:rgb(190 18 60 / var(--tw-text-opacity))!important}.text-slate-400{--tw-text-opacity: 1 !important;color:rgb(148 163 184 / var(--tw-text-opacity))!important}.text-slate-400\/70{color:#94a3b8b3!important}.text-slate-500{--tw-text-opacity: 1 !important;color:rgb(100 116 139 / var(--tw-text-opacity))!important}.text-white{--tw-text-opacity: 1 !important;color:rgb(255 255 255 / var(--tw-text-opacity))!important}.opacity-20{opacity:.2!important}.opacity-50{opacity:.5!important}.shadow{--tw-shadow: 0 1px 3px 0 rgb(0 0 0 / .1), 0 1px 2px -1px rgb(0 0 0 / .1) !important;--tw-shadow-colored: 0 1px 3px 0 var(--tw-shadow-color), 0 1px 2px -1px var(--tw-shadow-color) !important;box-shadow:var(--tw-ring-offset-shadow, 0 0 #0000),var(--tw-ring-shadow, 0 0 #0000),var(--tw-shadow)!important}.shadow-2xl{--tw-shadow: 0 25px 50px -12px rgb(0 0 0 / .25) !important;--tw-shadow-colored: 0 25px 50px -12px var(--tw-shadow-color) !important;box-shadow:var(--tw-ring-offset-shadow, 0 0 #0000),var(--tw-ring-shadow, 0 0 #0000),var(--tw-shadow)!important}.shadow-inner{--tw-shadow: inset 0 2px 4px 0 rgb(0 0 0 / .05) !important;--tw-shadow-colored: inset 0 2px 4px 0 var(--tw-shadow-color) !important;box-shadow:var(--tw-ring-offset-shadow, 0 0 #0000),var(--tw-ring-shadow, 0 0 #0000),var(--tw-shadow)!important}.shadow-lg{--tw-shadow: 0 10px 15px -3px rgb(0 0 0 / .1), 0 4px 6px -4px rgb(0 0 0 / .1) !important;--tw-shadow-colored: 0 10px 15px -3px var(--tw-shadow-color), 0 4px 6px -4px var(--tw-shadow-color) !important;box-shadow:var(--tw-ring-offset-shadow, 0 0 #0000),var(--tw-ring-shadow, 0 0 #0000),var(--tw-shadow)!important}.shadow-md{--tw-shadow: 0 4px 6px -1px rgb(0 0 0 / .1), 0 2px 4px -2px rgb(0 0 0 / .1) !important;--tw-shadow-colored: 0 4px 6px -1px var(--tw-shadow-color), 0 2px 4px -2px var(--tw-shadow-color) !important;box-shadow:var(--tw-ring-offset-shadow, 0 0 #0000),var(--tw-ring-shadow, 0 0 #0000),var(--tw-shadow)!important}.filter{filter:var(--tw-blur) var(--tw-brightness) var(--tw-contrast) var(--tw-grayscale) var(--tw-hue-rotate) var(--tw-invert) var(--tw-saturate) var(--tw-sepia) var(--tw-drop-shadow)!important}.transition{transition-property:color,background-color,border-color,text-decoration-color,fill,stroke,opacity,box-shadow,transform,filter,-webkit-backdrop-filter!important;transition-property:color,background-color,border-color,text-decoration-color,fill,stroke,opacity,box-shadow,transform,filter,backdrop-filter!important;transition-property:color,background-color,border-color,text-decoration-color,fill,stroke,opacity,box-shadow,transform,filter,backdrop-filter,-webkit-backdrop-filter!important;transition-timing-function:cubic-bezier(.4,0,.2,1)!important;transition-duration:.15s!important}.transition-all{transition-property:all!important;transition-timing-function:cubic-bezier(.4,0,.2,1)!important;transition-duration:.15s!important}.transition-transform{transition-property:transform!important;transition-timing-function:cubic-bezier(.4,0,.2,1)!important;transition-duration:.15s!important}.duration-200{transition-duration:.2s!important}.ease-in-out{transition-timing-function:cubic-bezier(.4,0,.2,1)!important}#gridmanContainer .flex-c{display:flex;flex-direction:column}#gridmanContainer .absolute-center-t{position:absolute;top:50%;left:50%;--tw-translate-x: -50%;--tw-translate-y: -50%;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))}#gridmanContainer .absolute-center-m{position:absolute;top:-100%;bottom:-100%;left:-100%;right:-100%;margin:auto}@keyframes rotate{to{transform:rotate(1turn)}}#gridmanContainer .bg-stripes-pink{background:repeating-linear-gradient(45deg,#c026d31a,#c026d31a 13px,#0000 13px,#0000 26px)!important}.hover\:scale-103:hover{--tw-scale-x: 1.03 !important;--tw-scale-y: 1.03 !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.hover\:bg-rose-500\/10:hover{background-color:#f43f5e1a!important}.hover\:bg-rose-500\/20:hover{background-color:#f43f5e33!important}.hover\:opacity-0:hover{opacity:0!important}.hover\:opacity-90:hover{opacity:.9!important}.\[\&\:\:placeholder\]\:text-slate-300::-moz-placeholder{--tw-text-opacity: 1 !important;color:rgb(203 213 225 / var(--tw-text-opacity))!important}.\[\&\:\:placeholder\]\:text-slate-300::placeholder{--tw-text-opacity: 1 !important;color:rgb(203 213 225 / var(--tw-text-opacity))!important}.\[\&\>div\:not\(\.Corner\)\]\:h-fit>div:not(.Corner){height:-moz-fit-content!important;height:fit-content!important}.\[\&\>div\]\:px-1>div{padding-left:4px!important;padding-right:4px!important}.\[\&\>div\]\:text-xs>div{font-size:12px!important}.\[\&\>div\]\:leading-tight>div{line-height:1.25!important}.\[\&\>div\]\:text-white>div{--tw-text-opacity: 1 !important;color:rgb(255 255 255 / var(--tw-text-opacity))!important}.\[\&\>div\]\:hover\:bg-slate-800:hover>div{--tw-bg-opacity: 1 !important;background-color:rgb(30 41 59 / var(--tw-bg-opacity))!important}.\[\&\>svg\]\:hover\:-translate-x-0\.5:hover>svg{--tw-translate-x: -2px !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}.\[\&\>svg\]\:hover\:translate-x-0\.5:hover>svg{--tw-translate-x: 2px !important;transform:translate(var(--tw-translate-x),var(--tw-translate-y)) rotate(var(--tw-rotate)) skew(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y))!important}</style><div id="gridmanContainer"><div class="z-[999999] fixed flex-c top-4 gap-1 text-white w-40 h-fit items-end right-4 "></div></div></template></div></body><div mx-name="view-image-info"><template shadowrootmode="closed">
    <style> #view_info_box{all:initial;position:fixed;z-index:999999999998;width:450px;height:295px;background:#293742;border-radius:3px;right:0;top:0;box-sizing:border-box;box-shadow:0 0 12px -2px rgba(32,32,32,.5);color:#d3d3d3;display:none;font-size:14px;padding-top:5px;font-family:Avenir,Helvetica,Arial,sans-serif}#view_info_box>div{clear:both;padding:5px 10px;width:auto!important;overflow:hidden;background:inherit;display:flex}#view_info_box .attr_name{width:90px;height:25px;line-height:25px}#view_info_box .attr_value{flex:1;text-align:center;background:#475f72;height:25px;line-height:25px;border-radius:2px;color:#ffffff}#view_info_box #address_link a:link{color:#ffffff}#view_info_box #address_link a:visited{color:#ffffff}#view_info_box #close_button{flex:1;text-align:center;padding-top:10px}#view_info_box .info_header{flex:1;text-align:center;font-weight:bold}#view_info_box .input_link{display:flex;height:22px;line-height:22px;align-content:center;width:100%;border:none;-webkit-appearance:none;outline:none;-webkit-tap-highlight-color:rgba(0,0,0,0);background:#475f72;color:#ffffff}.spinner{height:10px;width:10px;border:4px solid #d1d5db;border-top-color:#3b82f6;border-radius:50%;margin:0 auto;animation:spinner 800ms linear infinite;margin-top:3px}@keyframes spinner{from{transform:rotate(0deg)}to{transform:rotate(360deg)}}</style>
    <div id="view_info_box">
        <div title="header"><div class="info_header">View Image Info</div></div>

        <div title="Address" id="address_area">
            <div class="attr_name">Address:</div>
            <div class="attr_value" id="address_link" style="white-space: nowrap;overflow: hidden;text-overflow: ellipsis;padding:0 5px;"></div>
            <a target="_blank" class="address_link_svg" href="https://www.baidu.com/" style="color: #ffffff;display:flex;align-items:center;padding-left:5px;">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
            </a>

        </div>

        <div title="Original image dimensions (W x H pixels)">
            <div class="attr_name">Dimensions:</div>
            <div class="attr_value" id="dimensions_pix"></div>
        </div>

        <div title="Displayed image dimensions (W x H pixels)">
            <div class="attr_name">Displayed:</div>
            <div class="attr_value" id="displayed_pix"></div>
        </div>

        <div title="Alt Title">
            <div class="attr_name">Alt / Title</div>
            <div class="attr_value" id="file_desc" style="white-space: nowrap;overflow: hidden;text-overflow: ellipsis;padding: 0px 5px;"></div>
        </div>

        <div title="File Type">
            <div class="attr_name">File Type:</div>
            <div class="attr_value" id="file_type"><div class="spinner"></div></div>
        </div>

        <div title="File Size">
            <div class="attr_name">File Size:</div>
            <div class="attr_value" id="file_size"><div class="spinner"></div></div>
        </div>

        <div title="Close">
            <div id="close_button"><a id="close_link" style="margin-left:20px;padding:3px 20px;color:#ffffff;text-decoration:none;background: #475f72;border-radius:3px;" href="javascript:">Close</a></div>
        </div>
    </div>
    </template></div></html>